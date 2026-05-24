package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

var recvPort int = 49152

// all the local IPs
func GetAllIPs() {
	// 1. Get the list of interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error fetching interfaces:", err)
		return
	}

	for _, iface := range interfaces {
		// Print interface name (e.g., eth0, lo, wlan0)
		fmt.Printf("Interface: %s\n", iface.Name)

		// 2. Get addresses specifically for THIS interface
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("  Error getting addresses: %v\n", err)
			continue
		}

		// 3. List all IPs bound to it
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			//  Get the mask sizing (ones = current prefix, bits = total size 32 or 128)
			// ones, bits := ipNet.Mask.Size()

			fmt.Printf("  -> IP Address: %s | MASK: %s\n", addr.String(), ipNet.Mask)
		}
		fmt.Println() // Empty line for readability
	}
}

func GetIPToUse() (string, error) {
	Logger.Info("Finding temporary IPv6")

	// Query the system routing binary for the IP address attributes
	cmd := exec.Command("ip", "-6", "addr", "show", "dev", "wlp44s0")
	output, err := cmd.Output()
	if err != nil {
		Logger.Error("Error executing command", "error", err)
		// fmt.Println("Error executing command:", err)
		return "", errors.New("Something Went Wrong. Make sure it has admin privileges")
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Clean line whitespace and search for the IP definitions
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet6 ") {
			fields := strings.Fields(line)
			ipAndMask := fields[1]

			if strings.Contains(line, "temporary") {
				return ipAndMask, nil
			}

			// Evaluate system flags tagged at the end of the line
			// if strings.Contains(line, "temporary") {
			// 	fmt.Printf("[TEMPORARY PRIVACY IP] -> %s\n", ipAndMask)
			// } else if strings.Contains(line, "mngtmpaddr") || strings.Contains(line, "noprefixroute") {
			// 	fmt.Printf("[STABLE GLOBAL IP]    -> %s\n", ipAndMask)
			// } else if strings.Contains(line, "scope link") {
			// 	fmt.Printf("[LINK-LOCAL IP]       -> %s\n", ipAndMask)
			// }
		}
	}

	Logger.Error("Temporary IPv6 Not Found. Either enable that or change the browser setting and restart.")
	return "", errors.New("Temporary IPv6 Not Found. Either enable that or change the browser setting and restart.")
}

func generateTLSConfig() (*tls.Config, error) {
	// self-signed TLS certificate for QUIC (QUIC requires TLS 1.3)
	// not for production use
	Logger.Info("Generating Certificate")

	// Generate an RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		Logger.Error("Failed to generate private key", "error", err)
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Define the certificate template and create certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"QUO In-Memory Temporary Cert"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), // Valid for 24 hours
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		Logger.Error("Failed to create certificate", "error", err)
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode the certificate and key into PEM format and load it in-memory
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		Logger.Error("Failed to load x509 key pair", "error", err)
		return nil, fmt.Errorf("failed to load x509 key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}

func recv(listener *quic.Listener) {
	for {
		// Accept connection: This blocks until the QUIC/TLS handshake finishes
		conn, err := listener.Accept(context.Background())
		if err != nil {
			Logger.Error("Listener did not accept connection", "error", err)
			continue
		}

		Logger.Info("New connection made")

		// Pass the established connection to a handler
		go handleConn(conn)
	}
}
func handleConn(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			var appErr *quic.ApplicationError
			if errors.As(err, &appErr) {
				Logger.Info("Client closed connection", "code", appErr.ErrorCode)
				return
			}
			Logger.Error("Failed to accept stream", "error", err)
			return
		}
		defer stream.Close()

		Logger.Info("New stream accepted")

		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			Logger.Error("Read error", "error", err)
			return
		}

		Logger.Info("Received data successfully!", "content", string(buf[:n]))

		// Send ack so the client knows it's safe to close
		_, err = stream.Write([]byte("ACK"))
		if err != nil {
			Logger.Error("Write ack error", "error", err)
		}
	}
}

func Receiver() {
	if !Settings.Receiver {
		Logger.Info("Receiver is disabled")
		fmt.Println("Receiver is disabled")
		ReceiverStarted = false
		return
	}

	Logger.Info("Starting Receiver", "port", recvPort)

	// Generate standard TLS configuration required by QUIC
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		fmt.Print(err)
	}

	// Start the QUIC Listener on a UDP port
	listener, err := quic.ListenAddr("localhost:49152", tlsConfig, nil)
	if err != nil {
		Logger.Error("Listener did not initialize", "error", err)
		fmt.Print(err)
	}

	go recv(listener)
}

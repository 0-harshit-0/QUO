package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

var recvPort int = 49152

// Instead of a single net.UDPConn, we use a QUIC Transport and Listener
var transport *quic.Transport
var listener *quic.Listener

func GetLocalIPs() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

// Generates an ephemeral in-memory TLS certificate for the QUIC listener
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"p2p-sync"},
	}
}

// createTransport creates the underlying UDP socket and wraps it in a QUIC Transport
func createTransport() *quic.Transport {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: recvPort,
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	return &quic.Transport{
		Conn: udpConn,
	}
}

// Send QUIC stream payload
func send(dstPort int, dstAddr string, msg string) {
	ip := net.ParseIP(dstAddr)
	if ip == nil {
		Logger.Error("Invalid IPv4 address")
		return
	}

	addr := &net.UDPAddr{
		IP:   ip,
		Port: dstPort,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// InsecureSkipVerify is required because nodes are using self-signed certs
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"p2p-sync"},
	}

	// Dial using the shared Transport to maintain the same origin port
	conn, err := transport.Dial(ctx, addr, tlsConf, nil)
	if err != nil {
		Logger.Error("Not able to establish QUIC connection", "error", err)
		return
	}
	defer conn.CloseWithError(0, "done")

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		Logger.Error("Failed to open stream", "error", err)
		return
	}
	defer stream.Close()

	Logger.Info("Sending a packet...", "port", dstPort, "addr", dstAddr, "data", msg)

	_, err = stream.Write([]byte(msg))
	if err != nil {
		Logger.Error("Not able to send data on stream", "error", err)
	}
}

// Accept incoming QUIC connections
func recv() {
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			Logger.Error("Error accepting QUIC connection", "error", err)
			return
		}

		// Change: Use *quic.Conn instead of quic.Connection
		go func(c *quic.Conn) {
			for {
				stream, err := c.AcceptStream(context.Background())
				if err != nil {
					return
				}
				// stream here is already a *quic.Stream
				go handleStream(stream, c.RemoteAddr())
			}
		}(conn)
	}
}

// Handle an individual incoming stream
func handleStream(stream *quic.Stream, addr net.Addr) {
	defer stream.Close()

	// Read data from the stream
	buf, err := io.ReadAll(stream)
	if err != nil {
		Logger.Error("Error reading stream data", "error", err)
		return
	}

	var recvFromIP string
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		recvFromIP = udpAddr.IP.String()
	}

	recvData := string(buf)

	Logger.Info("Received a packet...",
		"port", recvPort,
		"addr", recvFromIP,
		"data_size", len(buf),
	)

	processRecvData(recvFromIP, recvData)
}

func processRecvData(ip string, data string) {
	substrings := strings.Split(data, ",")

	if len(substrings) == 0 {
		return
	}

	if strings.TrimSpace(substrings[0]) == "1" && Settings.AllowSync {
		var nodes []string
		for _, n := range AllNodes {
			nodes = append(nodes, fmt.Sprintf("%s:%d", n.Addr, n.Port))
		}

		payload := "n" + "," + strings.Join(nodes, ",") + "," + "0"
		// Note: No longer passing the connection; send() uses the global transport
		send(recvPort, ip, payload)
		return
	}

	if strings.TrimSpace(substrings[0]) == "n" && len(substrings) > 2 {
		for i := 1; i < len(substrings)-1; i++ {
			values := strings.Split(substrings[i], ":")

			if len(values) != 2 {
				continue
			}

			port, err := strconv.Atoi(values[1])
			if err != nil {
				continue
			}

			UpdateNodes(values[0], port)
		}

		SaveNodes()
	}
}

func RecvFrom() bool {
	if !Settings.Receiver {
		fmt.Println("Receiver is disabled")
		ReceiverStarted = false
		return false
	}

	transport = createTransport()

	var err error
	listener, err = transport.Listen(generateTLSConfig(), nil)
	if err != nil {
		panic(err)
	}

	go recv()

	Logger.Info("QUIC Receiver started", "port", recvPort)
	ReceiverStarted = true

	return true
}

func SyncNodes() {
	if !Settings.Receiver {
		fmt.Println("Enable the Receiver first")
		return
	}

	fmt.Println("Syncing...")

	for _, n := range AllNodes {
		send(n.Port, n.Addr, "1")
	}
}

func Cleanup() {
	if listener != nil {
		listener.Close()
	}
	if transport != nil {
		transport.Close()
	}
}

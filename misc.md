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


---------------------

You've run into the "Interface-to-Struct" transition. Since **v0.53.0**, the `quic-go` library swapped its main interfaces for concrete structs.

In **v0.59.0**:
1. `quic.Connection` was removed; it is now `*quic.Conn`.
2. `quic.Stream` is no longer an interface; it is a `*quic.Stream` struct.
3. Because they are structs, you must pass them as **pointers** (`*`) for them to satisfy standard interfaces like `io.Reader`.

### The Fixed Code for v0.59.0

Here are the specific blocks you need to update to clear those two errors:

#### 1. Fix the `recv` Loop
Update the goroutine signature to use `*quic.Conn`.
```go
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
```

#### 2. Fix the `handleStream` Signature
Update the parameter to use a pointer `*quic.Stream`. This allows `io.ReadAll` to see the `Read` method.
```go
// Change: Add the * to quic.Stream
func handleStream(stream *quic.Stream, addr net.Addr) {
	defer stream.Close()

	// Now io.ReadAll will work because *quic.Stream implements io.Reader
	buf, err := io.ReadAll(stream)
	if err != nil {
		Logger.Error("Error reading stream data", "error", err)
		return
	}
    
    // ... rest of your logic
}
```

### Why this happened
In older versions, `quic.Stream` was an interface, so `stream` (the variable) worked everywhere. In the new version, the `Read` and `Write` methods are defined on the **pointer** to the struct. 

By changing your function signature to `*quic.Stream`, you ensure you're passing the pointer, which satisfies the `io.Reader` requirement for `io.ReadAll(stream)`.

-------------------------

Here is a line-by-line breakdown of the QUIC implementation, explaining how it differs from standard UDP and why each part is necessary.

### 1. TLS Configuration (The "Handshake" Setup)
QUIC cannot run without encryption. These lines set up the security layer.

```go
func generateTLSConfig() *tls.Config {
    // Generates a new RSA private/public key pair (2048-bit).
    key, _ := rsa.GenerateKey(rand.Reader, 2048)
    
    // A basic certificate template. QUIC requires a serial number.
    template := x509.Certificate{SerialNumber: big.NewInt(1)}
    
    // Create the certificate in DER (binary) format using our key.
    certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
    
    // Convert the key and certificate into PEM (text) format for the TLS library.
    keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
    certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

    // Create a TLS certificate object from the PEM data.
    tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)

    return &tls.Config{
        Certificates: []tls.Certificate{tlsCert},
        // NextProtos (ALPN) is a string ID that both side must agree on
        // to ensure they are speaking the same application protocol.
        NextProtos: []string{"p2p-sync"},
    }
}
```

---

### 2. The Transport (The Socket Wrapper)
In QUIC, the "Transport" is the manager of the actual UDP socket.

```go
func createTransport() *quic.Transport {
    addr := &net.UDPAddr{IP: net.IPv4zero, Port: recvPort}
    
    // Create a standard UDP socket (same as your original code).
    udpConn, _ := net.ListenUDP("udp", addr)

    // Wrap that UDP socket in a QUIC Transport. 
    // This allows us to use one port for both listening and dialing.
    return &quic.Transport{Conn: udpConn}
}
```

---

### 3. Sending Data (The Client Role)
Unlike UDP, which just "fires and forgets," QUIC must establish a connection and a stream.

```go
func send(dstPort int, dstAddr string, msg string) {
    addr := &net.UDPAddr{IP: net.ParseIP(dstAddr), Port: dstPort}

    tlsConf := &tls.Config{
        InsecureSkipVerify: true,         // We don't have real CA certs, so skip validation.
        NextProtos:         []string{"p2p-sync"},
    }

    // Connect to the peer. Dial returns a *quic.Conn.
    conn, _ := transport.Dial(context.Background(), addr, tlsConf, nil)
    defer conn.CloseWithError(0, "done") // Cleanly shut down the connection when finished.

    // Open an outgoing stream. QUIC allows multiple streams on one connection; 
    // here we just use one to send our message.
    stream, _ := conn.OpenStreamSync(context.Background())
    defer stream.Close()

    // Write the message to the stream.
    stream.Write([]byte(msg))
}
```

---

### 4. Receiving Data (The Server Role)
This is a two-step process: Accept the **Connection**, then accept the **Stream**.

```go
func recv() {
    for {
        // Wait for a peer to initiate a connection.
        conn, _ := listener.Accept(context.Background())

        // Handle the connection in a goroutine so we don't block other peers.
        go func(c *quic.Conn) {
            for {
                // Inside a connection, wait for the peer to open a stream.
                stream, err := c.AcceptStream(context.Background())
                if err != nil { return } // Exit if connection is lost.
                
                // Process the individual stream.
                go handleStream(stream, c.RemoteAddr())
            }
        }(conn)
    }
}

func handleStream(stream *quic.Stream, addr net.Addr) {
    defer stream.Close()

    // io.ReadAll reads until the sender closes the stream.
    // This replaces your manual buffer [2048]byte.
    buf, _ := io.ReadAll(stream)

    // Extract the IP from the UDP address to pass to your processing logic.
    recvFromIP := addr.(*net.UDPAddr).IP.String()
    
    processRecvData(recvFromIP, string(buf))
}
```

---

### Summary of Differences
| Feature | Original UDP | `quic-go` |
| :--- | :--- | :--- |
| **Reliability** | Unreliable (packets can drop) | Reliable (retransmits lost data) |
| **Ordering** | Out of order | Streams are perfectly ordered |
| **Security** | Plaintext | Mandatory TLS 1.3 Encryption |
| **Flow Control** | None (can overwhelm receiver) | Built-in (prevents flooding) |
| **Addressing** | `ReadFromUDP` gives the address | `c.RemoteAddr()` gives the address |

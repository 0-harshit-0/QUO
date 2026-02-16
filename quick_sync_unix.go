//go:build linux || darwin

package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

var recvPort int = 54321
var conn *net.UDPConn


func InitWinsock() {
	// no-op on non-Windows
}

func CleanupWinsock() {
	// no-op on non-Windows
}

func GetLocalIPs() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// skip down or loopback interfaces
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



// Create and bind UDP socket
func createUDPSocket() *net.UDPConn {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: recvPort,
	}

	c, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	return c
}

// Send UDP packet
func send(conn *net.UDPConn, dstPort int, dstAddr string, msg string) {
	ip := net.ParseIP(dstAddr)
	if ip == nil {
		Logger.Error("Invalid IPv4 address")
		return
	}

	addr := &net.UDPAddr{
		IP:   ip,
		Port: dstPort,
	}

	Logger.Info("Sending a packet...", "port", dstPort, "addr", dstAddr, "data", msg)

	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		Logger.Error("Send error:", err)
	}
}

// Receive loop
func recv(conn *net.UDPConn) {
	buf := make([]byte, 2048)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			Logger.Error("Recv error:", err)
			return
		}

		recvFromIP := addr.IP.String()
		recvData := string(buf[:n])

		Logger.Info("Received a packet...",
			"port", recvPort,
			"addr", recvFromIP,
			"data_size", n,
		)

		processRecvData(recvFromIP, recvData)
	}
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

		SendTo(ip, "n", strings.Join(nodes, ","), "0")
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

func CheckActive() {
	if !Settings.Receiver {
		fmt.Println("Enable the Receiver first")
		return
	}

	fmt.Println("Syncing...")

	for _, n := range AllNodes {
		send(conn, n.Port, n.Addr, "1")
	}
}

func RecvFrom() bool {
	if !Settings.Receiver {
		fmt.Println("Receiver is disabled")
		return false
	}

	conn = createUDPSocket()

	Logger.Info("Receiver started", "port", recvPort)

	go recv(conn)

	return true
}

func SendTo(dstAddr string, datatype string, data string, syncFlag string) error {
	payload := datatype + "," + data + "," + syncFlag
	send(conn, recvPort, dstAddr, payload)
	return nil
}

func Cleanup() {
	if conn != nil {
		conn.Close()
	}
}

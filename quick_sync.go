//go:build windows

package main


import (
	"golang.org/x/sys/windows"
	"fmt"
	"net"
	"strings"
)


type Node struct {
	Addr         string `json:"addr"`
	Port         int `json:"port"`
	CheckedCount int    `json:"checked_count"`
	Type         string `json:"type"`
}

var max_checked_count int = 6
var recv_port int = 54321


// goes to the main thread
func InitWinsock() {
	// Initialize Winsock
	var wsa windows.WSAData
	err := windows.WSAStartup(uint32(0x202), &wsa)
	if err != nil {
		panic(err)
	}
}
func CleanupWinsock() {
	windows.Closesocket(WinSocket)
	windows.WSACleanup()
}


func createUDPSocket() windows.Handle {
	sock, err := windows.Socket(
		windows.AF_INET,
		windows.SOCK_DGRAM,
		windows.IPPROTO_UDP,
	)
	if err != nil {
		panic(err)
	}
	// defer windows.Closesocket(sock)
	return sock
}

func send(sock windows.Handle, dst_port int, dst_addr string, msg string) {
	ip := net.ParseIP(dst_addr)
	if ip == nil {
		panic("invalid IP address")
	}

	ip4 := ip.To4()
	if ip4 == nil {
		panic("not an IPv4 address")
	}

	addr := &windows.SockaddrInet4{
		Port: dst_port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}
	err := windows.Sendto(sock, []byte(msg), 0, addr)
	if err != nil {
		panic(err)
	}
}

func recv(sock windows.Handle) {
	addr := &windows.SockaddrInet4{
		Port: recv_port,
		Addr: [4]byte{0, 0, 0, 0}, // INADDR_ANY
	}

    err = windows.Bind(sock, addr)
	if err != nil {
		panic(err)
	}

	fmt.Print("Listening on UDP port %s\n", recv_port)

	buf := make([]byte, 2048)

	for {
		n, from, err := windows.Recvfrom(sock, buf, 0)
		if err != nil {
			fmt.Println("recv error:", err)
			continue
		}

		fromAddr := from.(*windows.SockaddrInet4)
		fmt.Printf(
			"Received %d bytes from %d.%d.%d.%d:%d: %s\n",
			n,
			fromAddr.Addr[0],
			fromAddr.Addr[1],
			fromAddr.Addr[2],
			fromAddr.Addr[3],
			fromAddr.Port,
			string(buf[:n]),
		)
	}
}


var WinSocket windows.Handle = createUDPSocket()

func CheckActive() {
    fmt.Println("Syncing...")

    // ask all the nodes if they are live

    nodes, err := ReadJson[[]Node](CacheDir+"/nodes.json")
    if err != nil {
        return
    }

    for _, n := range nodes {
    	if n.CheckedCount < max_checked_count {
			send(WinSocket, n.Port, n.Addr, "1")
    	}
	}
}

func RecNodes() {
    fmt.Println("Receiving Nodes...")

}

func SendNodes(receive string) {
    fmt.Println("Sending Nodes...")

    nodes, err := ReadJson[[]Node](CacheDir+"/nodes.json")
    if err != nil {
        return
    }

    addrs := make([]string, 0, len(nodes)) // lnegth: 0 | capacity: len(nodes)
    for _, n := range nodes {
    	if n.CheckedCount < (max_checked_count/2) {
			addrs = append(addrs, n.Addr)
    	}
	}
	addrs = append(addrs, receive)// append the flag 0 or 1

	send(WinSocket, 54321, "127.0.0.1", strings.Join(addrs, ","))
}


//go:build windows

package main


import (
	"golang.org/x/sys/windows"
	"fmt"
	"strings"
)


type Node struct {
	Addr         string `json:"addr"`
	CheckedCount int    `json:"checked_count"`
	Type         string `json:"type"`
}

var max_checked_count int = 6
var wsa windows.WSAData


func InitWinsock() {
	// Initialize Winsock
	err := windows.WSAStartup(uint32(0x202), &wsa)
	if err != nil {
		panic(err)
	}
}
func CleanupWinsock() {
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

func send(sock windows.Handle, msg string) {
	addr := &windows.SockaddrInet4{
		Port: 54321,
		Addr: [4]byte{127, 0, 0, 1},
	}
	err := windows.Sendto(sock, []byte(msg), 0, addr)
	if err != nil {
		panic(err)
	}
}


sock := createUDPSocket()
func StartHandshake() {
    fmt.Println("Syncing...")

	send(sock, "1")
}
func SendNodes(receive string) {
    fmt.Println("Sending Nodes...")

    nodes, err := ReadJson[[]Node](CacheDir+"/nodes.json")
    if err != nil {
        return
    }

    addrs := make([]string, 0, len(nodes)) // lnegth: 0 | capacity: len(nodes)
    for _, n := range nodes {
    	if n.CheckedCount < (max_checked_count/2):
			addrs = append(addrs, n.Addr)
	}
	addrs.append(receive) // append the flag 0 or 1

	send(sock, strings.Join(addrs, ","))
}
func RecNodes() {
    fmt.Println("Receiving Nodes...")

}
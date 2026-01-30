//go:build windows

package main


import (
	"golang.org/x/sys/windows"
	"fmt"
)

func StartBeacon() {
	fmt.Println("hey")
	// Initialize Winsock
	var wsa windows.WSAData
	err := windows.WSAStartup(uint32(0x202), &wsa)
	if err != nil {
		panic(err)
	}
	defer windows.WSACleanup()

	// Create UDP socket
	sock, err := windows.Socket(
		windows.AF_INET,
		windows.SOCK_DGRAM,
		windows.IPPROTO_UDP,
	)
	if err != nil {
		panic(err)
	}
	defer windows.Closesocket(sock)

	// Destination address
	addr := &windows.SockaddrInet4{
		Port: 54321,
		Addr: [4]byte{127, 0, 0, 1},
	}

	// Data to send
	data := []byte("hello udp")

	// Send packet
	err = windows.Sendto(
		sock,
		data,
		0,
		addr,
	)
	if err != nil {
		panic(err)
	}
}

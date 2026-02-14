//go:build windows

package main


import (
	"golang.org/x/sys/windows"
	"fmt"
	"net"
	"strings"
)


type nodeJson struct {
	Addr         string `json:"addr"`
	Port         int    `json:"port"`
	CheckedCount int    `json:"checked_count"`
}


var allNodes []nodeJson
var recvPort int = 54321

var winSocket windows.Handle = createUDPSocket()
// received_data := make(map[string]string)


// goes to the main thread
func InitWinsock() {
	// Initialize Winsock
	var wsa windows.WSAData

	err := windows.WSAStartup(uint32(0x202), &wsa)
	if err != nil {
		panic(err)
	}
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
		Port: recvPort,
		Addr: [4]byte{0, 0, 0, 0}, // INADDR_ANY
	}

    err := windows.Bind(sock, addr)
	if err != nil {
		panic(err)
	}

	// fmt.Print("Listening on UDP port %s\n", recvPort)

	buf := make([]byte, 2048)

	for {
		n, from, err := windows.Recvfrom(sock, buf, 0)
		if err != nil {
			fmt.Println("recv error:", err)
			return
		}

		fromAddr, ok := from.(*windows.SockaddrInet4)
		if !ok {
			fmt.Println("not an IPv4 address")
			return
		}

		// fmt.Printf(
		// 	"Received %d bytes from %d.%d.%d.%d:%d: %s\n",
		// 	n,
		// 	fromAddr.Addr[0],
		// 	fromAddr.Addr[1],
		// 	fromAddr.Addr[2],
		// 	fromAddr.Addr[3],
		// 	fromAddr.Port,
		// 	string(buf[:n]),
		// )

		recvFromIP := fmt.Sprintf("%d.%d.%d.%d",
			fromAddr.Addr[0],
			fromAddr.Addr[1],
			fromAddr.Addr[2],
			fromAddr.Addr[3],
		)

		processRecvData(recvFromIP, string(buf[:n]))
	}
}

func processRecvData(ip string, data string) {
	substrings := strings.Split(data, ",")

	if len(substrings) > 0 && strings.TrimSpace(substrings[0]) == "1" && Settings.AllowSync {
		// asking for nodes to complete sync
	    var nodes []string
		for _, n := range allNodes {
			nodes = append(nodes, fmt.Sprintf("%s:%d", n.Addr, n.Port))
		}

		SendToNode(ip, "n", strings.Join(nodes, ","), "0")
	}

}


func LoadNodes(maxCount int) {
	nodes, err := ReadJson[[]nodeJson](ConfigDir+"/nodes.json")
    if err != nil {
        return
    }

    for _, n := range nodes {
    	if n.CheckedCount < maxCount {
    		allNodes = append(allNodes, n)
    	}
	}

	// return allNodes
}

func CheckActive() {
	if Settings.Receiver == false {
		fmt.Println("Enable the Receiver first")
		return
	}

    fmt.Println("Syncing...")

    // ask all the nodes if they are live
    for _, n := range allNodes {
		send(winSocket, n.Port, n.Addr, "1")
	}
}

func RecvFromNodes() bool {
	if Settings.Receiver == false {
		fmt.Println("Receiver is disabled")

		CleanupWinsock()
		return Settings.Receiver
	}

    fmt.Printf("Receiver started at port %s\n", recvPort)
    go recv(winSocket)

    return Settings.Receiver
}

func SendToNode(dst_addr string, datatype string, data string, sync_flag string) error {
	// datatype: n for nodes, w for webpages
	// sync_flag: 1 or 0
    
    // fmt.Println("Sending...")

    payload := datatype + "," + data + "," + sync_flag

	send(winSocket, recvPort, dst_addr, payload)

	return nil
}


func CleanupWinsock() {
	windows.Closesocket(winSocket)
	windows.WSACleanup()
}




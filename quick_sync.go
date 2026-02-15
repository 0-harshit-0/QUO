//go:build windows

package main


import (
	"golang.org/x/sys/windows"
	"os"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
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
		Logger.Error("UDP socket error", err)
	}

	// defer windows.Closesocket(sock)
	return sock
}


func send(sock windows.Handle, dst_port int, dst_addr string, msg string) {
	ip := net.ParseIP(dst_addr)
	if ip == nil {
		Logger.Error("Invalid IPv4 address")
		return
	}

	ip4 := ip.To4()
	if ip4 == nil {
		Logger.Error("Invalid IPv4 address")
		return
	}

	addr := &windows.SockaddrInet4{
		Port: dst_port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}

    Logger.Info("Sending a packet...", "port", dst_port, "addr", dst_addr, "data", msg)

	err := windows.Sendto(sock, []byte(msg), 0, addr)
	if err != nil {
		Logger.Error("Send error:", err)
		return
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
			Logger.Error("Recv error:", err)
			return
		}

		fromAddr, ok := from.(*windows.SockaddrInet4)
		if !ok {
			Logger.Error("Not an IPv4 address")
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
		recvData := string(buf[:n])

    	Logger.Info("Received a packet...", "port", recvPort, "addr", recvFromIP, "data_size", n)

		processRecvData(recvFromIP, recvData)
	}
}

func processRecvData(ip string, data string) {
	substrings := strings.Split(data, ",")

	if len(substrings) == 0 {
		return
	}

	if strings.TrimSpace(substrings[0]) == "1" && Settings.AllowSync {
		// asking for nodes to complete sync
	    var nodes []string
		for _, n := range allNodes {
			nodes = append(nodes, fmt.Sprintf("%s:%d", n.Addr, n.Port))
		}

		SendToNode(ip, "n", strings.Join(nodes, ","), "0")
		return
	}

	if strings.TrimSpace(substrings[0]) == "n" && len(substrings) > 2 {
		// receiving nodes, save them
		for i := 1; i < len(substrings)-1; i++ {
			values := strings.Split(substrings[i], ":")

			port, _ := strconv.Atoi(values[1])

		    newNode := nodeJson{
		    	Addr: values[0],
		    	Port: port,
		    	CheckedCount: 0,
		    }

		    allNodes = append(allNodes, newNode)
		}

		SaveNodes()
	}
}


func LoadNodes(maxCount int) {
    Logger.Info("Loading Nodes Config File")

	nodes, err := ReadJson[[]nodeJson](ConfigDir+"/nodes.json")
    if err != nil {
    	Logger.Error("Error loading nodes", err)
        return
    }

    for _, n := range nodes {
    	if n.CheckedCount < maxCount {
    		allNodes = append(allNodes, n)
    	}
	}

	// return allNodes
}

func SaveNodes() {
    path := ConfigDir+"/nodes.json"

    // write back to file
    out, err := json.MarshalIndent(allNodes, "", "  ")
    if err != nil {
    	Logger.Error("Error saving nodes", err)
        return
    }

    os.WriteFile(path, out, 0644)
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

    fmt.Printf("Receiver started at port %d\n", recvPort)
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




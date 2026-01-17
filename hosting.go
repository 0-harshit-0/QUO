package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
)


func Host(path string, name string, port uint16, completed chan bool) *http.Server {
	mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir(path)))

	addr := fmt.Sprintf(":%d", port)

    srv := &http.Server{
        Addr:    addr,
        Handler: mux,
    }

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        completed <- true
        return nil
    }

	fmt.Printf("Serving %s at http://localhost%s\n", name, addr)

    go func() {
        err := srv.Serve(ln)
        if err != nil && err != http.ErrServerClosed {
            log.Println("server error:", err)
        }
    }()

    completed <- true
    return srv
}
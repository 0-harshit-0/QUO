package main

import (
    "os"
    "fmt"
    "log"
    "net"
    "net/http"
)


func Host(path string, name string, port uint16) (*http.Server, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("path is not a directory: %s", path)
    }

	mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir(path)))

	addr := fmt.Sprintf(":%d", port)

    srv := &http.Server{
        Addr:    addr,
        Handler: mux,
    }

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }

	fmt.Printf("Serving %s at http://localhost%s\n", name, addr)

    go func() {
        err := srv.Serve(ln)
        if err != nil && err != http.ErrServerClosed {
            log.Println("server error:", err)
        }
    }()

    return srv, nil
}
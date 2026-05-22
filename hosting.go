package main

import (
	"context"
	"time"
	"fmt"
	"net"
	"net/http"
	"os"
)

func Host(path string, name string, port uint16) (*http.Server, error) {
	Logger.Info("Starting Server")

	info, err := os.Stat(path)
	if err != nil {
		Logger.Error("Trouble staring the Host", "error", err)
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
		Logger.Error("Trouble listening", "error", err)
		return nil, err
	}

	Logger.Info("Serving %s at http://localhost%s\n", name, addr)
	fmt.Printf("Serving %s at http://localhost%s\n", name, addr)

	go func() {
		err := srv.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			Logger.Error("Trouble staring the Host", "error", err)
		}
	}()

	return srv, nil
}
func CloseHost(t *tab) {
	if t.serving == true {
		Logger.Info("Closing", "port", t.port)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		err := t.server.Shutdown(ctx)
		if err != nil {
			Logger.Error("trouble shutting down", "error", err)
		}

		cancel() // defer it for gracefullness

		// reset
		t.server = nil
		t.serving = false
	}
}

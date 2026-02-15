package main

import (
	"log/slog"
	"os"
)


var Logger *slog.Logger


func NewLogger() (*slog.Logger, *os.File, error) {
	file, err := os.Create(CacheDir+"/app.log")
	// file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, nil, err
	}

	handler := slog.NewTextHandler(file, nil)
	Logger = slog.New(handler)
	return Logger, file, nil
}

package main

import (
	"fmt"
	"strings"
)

var Shutdown = make(chan struct{})
var ConfigDir string = "config"
var CacheDir string = "cache"
var ReceiverStarted bool = false

func BrowserState() {
	ips := GetLocalIPs()

	fmt.Printf("Current Tab: %d | Total tabs: %d | IP: %v | Receiver: %t\n", CurrentTabID, len(Tabs), ips, ReceiverStarted)
}
func QuitBrowser() {
	//  When the channel(pipe) is closed, it signals each tab and closes active host.
	//  After this it returns and the HelpMessage loop ends, which closes of the last goroutine and leads to a graceful shutdown
	close(Shutdown)
	Tabs = nil
}

func main() {
	_, file, err := NewLogger()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	Logger.Info("Starting Browser")

	// start a common winsock for the whole program
	InitWinsock()
	defer CleanupWinsock()

	// initial prints
	asciiName := `
------------------------ __ _ _   _  ___ -------------------------
----------------------- / _* | | | |/ _ \ ------------------------
---------------------- | (_| | |_| | (_) | -----------------------
----------------------- \__, |\__,_|\___/ ------------------------
-------------------------- |_| -----------------------------------
`
	fmt.Println(strings.ReplaceAll(asciiName, "-", " "))

	// load settings and files
	LoadSettings()
	LoadNodes()

	// the quick-sync receiver
	RecvFrom()

	// start the browser, by starting a tab, etc.
	NewTab()

	// show the browser state
	BrowserState()
	fmt.Println()

	// user I/O
	done := make(chan struct{})
	go func() {
		// ReadWebpagesFolder(0, 10)
		HelpMenu()
		close(done)
	}()
	<-done

	Logger.Info("Closing Browser")
}

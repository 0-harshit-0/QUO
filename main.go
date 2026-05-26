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
	ip, err := GetIPToUse()
	if err != nil {
		fmt.Printf("Error getting addresses: %v\n", err)
		return
	}

	fmt.Printf("Current Tab: %d | Total tabs: %d | Current IP: %v | Receiver Started: %t | Nodes: %d\n", CurrentTabID, len(Tabs), ip, ReceiverStarted, len(AllNodes))
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

	// initial prints
	asciiName := `------------------------ __ _ _   _  ___ -------------------------
----------------------- / _* | | | |/ _ \ ------------------------
---------------------- | (_| | |_| | (_) | -----------------------
----------------------- \__, |\__,_|\___/ ------------------------
-------------------------- |_| -----------------------------------
`
	fmt.Println(strings.ReplaceAll(asciiName, "-", " "))

	// load settings and files
	LoadSettings()
	LoadNodes()

	// the quick-sync
	CreateTransport()
	Receiver()

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

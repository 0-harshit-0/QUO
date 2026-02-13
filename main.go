package main

import (
    "fmt"
    "strings"
)



var Shutdown = make(chan struct{})
var ConfigDir string = "config"
var CacheDir string = "cache"
var RootWebpagesFolder string = "webpages" // root webpages folder



func BrowserState() {
    fmt.Printf("Current Tab: %d | Total tabs: %d\n", CurrentTabID, len(Tabs))
}
func QuitBrowser() {
    //  When the channel(pipe) is closed, it signals each tab and closes active host. After this it returns and the HelpMessage loop ends which results in the closure of the last goroutine and leads to a graceful shutdown
    close(Shutdown)
    Tabs = nil
}

func main() {
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
    fmt.Println(strings.ReplaceAll(asciiName, "-", " "), "\n")
    
    // start the browser, by starting a tab, etc.
    LoadSettings()
    NewTab(true)

    RecvFromNodes()
    BrowserState()
    fmt.Println()
    
    done := make(chan struct{})
    go func() {
        // ReadWebpagesFolder(0, 10)
        HelpMenu()
        close(done)
    }()
    <- done
}
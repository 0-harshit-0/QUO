package main

import (
    "fmt"
    "strings"
)



var Shutdown = make(chan struct{})
var CacheDir string = "cache"



func BrowserState() {
    fmt.Printf("Current Tab: %d | Total tabs: %d\n", CurrentTabID, len(Tabs))
}
func QuitBrowser() {
    close(Shutdown)
    Tabs = nil
}

func main() {
    asciiName := `
------------------------ __ _ _   _  ___ -------------------------
----------------------- / _* | | | |/ _ \ ------------------------
---------------------- | (_| | |_| | (_) | -----------------------
----------------------- \__, |\__,_|\___/ ------------------------
-------------------------- |_| -----------------------------------
`
    fmt.Println(strings.ReplaceAll(asciiName, "-", " "), "\n")
    
    // start the browser, by starting a tab
    NewTab(true)
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
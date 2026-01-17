package main

import (
    "fmt"
    "strings"
)



var Shutdown = make(chan struct{})



func BrowserState() {
    fmt.Printf("Cuurent Tab: %d | Total tabs: %d\n", CurrentTabID, len(Tabs))
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
        ReadWebpagesFolder()
        HelpMenu()
        close(done)
    }()
    <- done
}
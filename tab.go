package main

import (
    "fmt"
    "context"
    "time"
    // "strconv"
    "net/http"
)


type Command struct {
    Action string
    PageIndex  int
    Completed chan bool // flag to know when the command is ended
}
type tab struct {
    id      int
    port    uint16
    serving bool
    server  *http.Server
    command chan Command
}


var Tabs = make(map[int]*tab)
var CurrentTabID int = 0 // 0 id means theres no tab
var nextTabID int = 1

var portToUse uint16 = 49152
const maxPort = 65535

func (t *tab) run() {
    for {
        select {
            case cmd, ok := <-t.command:
                if !ok {
                    // channel closed
                    // select does not close the routine automatically
                    // for range closes the loop automatically
                    return
                }
                switch cmd.Action {
                    case "start_server":
                        // id, err := strconv.Atoi(cmd.Query)
                        // if err != nil {
                        //     fmt.Println("\nInvalid folder id\n")
                        //     continue
                        // }
                        if cmd.PageIndex < 0 || cmd.PageIndex > len(Webpages) {
                            fmt.Println("\nFolder not found")
                            cmd.Completed <- true
                            continue
                        }

                        // shutting down the old host
                        CloseHost(t)

                        if t.serving {//just a fail safe
                            fmt.Printf("Already serving at port: %d\n", t.port)
                            cmd.Completed <- true
                            continue
                        }

                        srv, err := Host(Webpages[cmd.PageIndex].Path, Webpages[cmd.PageIndex].Name, t.port)
                        if err != nil {
                            fmt.Println("Error: ", err)
                            cmd.Completed <- true
                            continue
                        }
                        // <- cmd.Completed
                        t.server = srv
                        t.serving = true

                        err = UpdateHistory(Webpages[cmd.PageIndex].Name)
                        if err != nil {
                            fmt.Println(err)
                        }

                        cmd.Completed <- true
                }
                
            case <-Shutdown:
                // close all host as well before shutdown
                CloseHost(t)
                return
        }
    }
}


func NewTab(nofmt bool) {
    if portToUse > maxPort {
        panic("ran out of available ports")
    }

    tab := &tab{
        id:      nextTabID,
        port:    portToUse,
        serving: false,
        server:  nil,
        command: make(chan Command),
    }
    go tab.run()

    Tabs[tab.id] = tab
    nextTabID++
    portToUse++

    if !nofmt {
        fmt.Printf("New tab (%d) opened\n", tab.id)
    }

    // switch the tab
    SwitchTab(tab.id, nofmt)
}

func SwitchTab(id int, nofmt bool) {
    _, ok := Tabs[id]
    if ok {
        CurrentTabID = id
        if !nofmt {
            fmt.Printf("Switched to %d", CurrentTabID)
        }
    } else {
        fmt.Println("Tab does not exist")
    }
}

func CloseTab(id int) {
    fmt.Printf("Closing: %d\n", id)

    tab, ok := Tabs[id];
    if ok {
        // closing the host inside the tab
        CloseHost(tab)

        // closing the tab
        close(tab.command)
        delete(Tabs, id)
        
        if len(Tabs) == 0 {
            fmt.Println("All tabs closed")
            CurrentTabID = 0
            return
        }

        // replacing the current tab id with the other one else it will point to a closed tab
        for key, _ := range Tabs {
            CurrentTabID = key
            fmt.Printf("Switched to: %d", CurrentTabID)
            break
        }
    } else {
        fmt.Println("No tab to close or something went wrong")
    }
}

func CloseHost(t *tab) {
    if t.serving == true {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

        err := t.server.Shutdown(ctx);
        if err != nil {
            fmt.Println("shutdown error:", err)
        }

        cancel() // defer it for gracefullness
        fmt.Printf("Closing port: %d\n", t.port)
        // t.server = nil
        t.serving = false
    }
}
// func CloseAllTabs() {
//     for id, tab := range Tabs {
//         close(tab.command)
//         delete(Tabs, id)
//     }
// }


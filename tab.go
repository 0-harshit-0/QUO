package main

import (
    "fmt"
    "context"
    "log"
    "time"
    // "strconv"
    "net/http"
)


type Command struct {
    Action string
    PageId  int
    Completed chan bool // flag to know when the command is ended
}
type Tab struct {
    id      int
    port    uint16
    serving bool
    server  *http.Server
    command chan Command
}


var Tabs = make(map[int]*Tab)
var CurrentTabID int = 0 // 0 id means theres no tab
var NextTabID int = 1

var portToUse uint16 = 49152
const maxPort = 65535

func (t *Tab) run() {
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
                        if cmd.PageId < 0 || cmd.PageId > len(Sites) {
                            log.Println("\nFolder not found")
                            cmd.Completed <- true
                            continue
                        }

                        // shutting down the old host
                        CloseHost(t)

                        if t.serving {//just a fail safe
                            log.Printf("Already serving at port: %d\n", t.port)
                            cmd.Completed <- true
                            continue
                        }

                        srv, err := Host(Sites[cmd.PageId].Path, Sites[cmd.PageId].Name, t.port)
                        if err != nil {
                            log.Print("Error: ", err)
                            cmd.Completed <- true
                            continue
                        }
                        // <- cmd.Completed
                        t.server = srv
                        t.serving = true
                        cmd.Completed <- true
                }
                
            case <-Shutdown:
                // close all host as well before shutdown
                CloseHost(t)
                return
        }
    }
}


func NewTab(noLog bool) {
    if portToUse > maxPort {
        panic("ran out of available ports")
    }

    tab := &Tab{
        id:      NextTabID,
        port:    portToUse,
        serving: false,
        server:  nil,
        command: make(chan Command),
    }
    go tab.run()

    Tabs[tab.id] = tab
    NextTabID++
    portToUse++

    if !noLog {
        fmt.Printf("New tab (%d) opened\n", tab.id)
    }

    // switch the tab
    SwitchTab(tab.id, noLog)
}

func SwitchTab(id int, noLog bool) {
    _, ok := Tabs[id]
    if ok {
        CurrentTabID = id
        if !noLog {
            fmt.Printf("Switched to %d", CurrentTabID)
        }
    } else {
        log.Println("Tab does not exist")
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
        log.Println("No tab to close or something went wrong")
    }
}

func CloseHost(t *Tab) {
    if t.serving == true {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

        err := t.server.Shutdown(ctx);
        if err != nil {
            log.Println("shutdown error:", err)
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


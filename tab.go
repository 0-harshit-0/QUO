package main

import (
	"fmt"

	// "strconv"
	"net/http"
)

type Command struct {
	Action    string
	PageIndex int
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
var newTabID int = 1
var portToUse uint16 = 49153

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
				Logger.Info("Open New Webpage")

				// id, err := strconv.Atoi(cmd.Query)
				// if err != nil {
				//     fmt.Println("\nInvalid folder id\n")
				//     continue
				// }

				if cmd.PageIndex < 0 || cmd.PageIndex > len(Webpages) || len(Webpages) == 0 {
					Logger.Error("Webpage not found")
					fmt.Println("Webpage not found")
					cmd.Completed <- true
					continue
				}

				// shutting down the old host
				CloseHost(t)

				if t.serving { //just a fail safe, if abve func returns error
					Logger.Error("Already serving", "port", t.port)
					fmt.Printf("Already serving at port: %d\n", t.port)
					cmd.Completed <- true
					continue
				}

				srv, err := Host(Webpages[cmd.PageIndex].Path, Webpages[cmd.PageIndex].Name, t.port)
				if err != nil {
					fmt.Println(err)
					cmd.Completed <- true
					continue
				}

				t.server = srv
				t.serving = true

				err = UpdateHistory(Webpages[cmd.PageIndex].Name)
				if err != nil {
					fmt.Println(err)
				}

				// <- cmd.Completed
				cmd.Completed <- true
			}

		case <-Shutdown:
			// close all host as well before shutdown
			CloseHost(t)
			return
		}
	}
}

func NewTab() {
	Logger.Info("Opening New tab", "tab", newTabID)

	// extra check
	if portToUse > maxPort {
		Logger.Error("Ran out of ports")
		panic("ran out of available ports")
	}

	tab := &tab{
		id:      newTabID,
		port:    portToUse,
		serving: false,
		server:  nil,
		command: make(chan Command),
	}
	go tab.run()
	Tabs[tab.id] = tab

	// switch the tab
	SwitchTab(tab.id)

	// Logger.Info("New tab opened", "tab", tab.id)

	newTabID++
	portToUse++
}

func SwitchTab(id int) {
	Logger.Info("Switching to", "tab", id)

	_, ok := Tabs[id]
	if ok {
		CurrentTabID = id
		// fmt.Printf("\nSwitched to tab: %d", CurrentTabID)
		// Logger.Info("Switched to", "tab", CurrentTabID)
	} else {
		Logger.Error("Tab does not exist")
		fmt.Println("Tab does not exist")
	}
}

func CloseTab(id int) {
	Logger.Info("Closing", "tab", id)

	tab, ok := Tabs[id]
	if ok {
		// closing the host inside the tab
		CloseHost(tab)

		// closing the tab
		close(tab.command)
		delete(Tabs, id)

		if len(Tabs) == 0 {
			CurrentTabID = 0
			Logger.Info("All tabs closed")
			return
		}

		// replacing the current tab id with the other one, else it will point to a closed tab
		for key, _ := range Tabs {
			CurrentTabID = key
			Logger.Info("Switched to", "tab", CurrentTabID)
			break
		}
	} else {
		Logger.Info("No tab to close")
		fmt.Println("No tab to close or something went wrong")
	}
}

// func CloseAllTabs() {
//     for id, tab := range Tabs {
//         close(tab.command)
//         delete(Tabs, id)
//     }
// }

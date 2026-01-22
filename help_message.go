package main

import (
    "fmt"
)


func SInput() (string) {
    var input string
    _, err := fmt.Scanln(&input)
    if err != nil {
        fmt.Println("Error reading input:", err)
        return ""
    }
    return input
}
func NInput() (int) {
    var input int
    _, err := fmt.Scanln(&input)
    if err != nil {
        fmt.Println("Error reading input:", err)
        return 0
    }

    return input
}

func HelpMenu() {
    msgs := [...]string{
        0: "Quit the browser",
        1: "Browser state",
        2: "Browser commands",
        3: "New tab",
        4: "Switch tab",
        5: "Close current tab",
        6: "Search",
        7: "Close Host",
    }
    rows := 3
    cols := (len(msgs) + rows - 1) / rows

    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            i := c*rows + r
            if i < len(msgs) {
                fmt.Printf("%-25s", fmt.Sprintf("%d - %s", i, msgs[i]))
            }
        }
        fmt.Println()
    }

    for {
        // input
        fmt.Print("\nEnter command: ")
        input := NInput()

        switch input {
            case 0:
                fmt.Println("ta! ta!")
                QuitBrowser()
                return
            case 1:
                BrowserState()
            case 2:
                for r := 0; r < rows; r++ {
                    for c := 0; c < cols; c++ {
                        i := c*rows + r
                        if i < len(msgs) {
                            fmt.Printf("%-25s", fmt.Sprintf("%d - %s", i, msgs[i]))
                        }
                    }
                    fmt.Println()
                }
            case 3:
                NewTab(false)
            case 4:
                fmt.Println("Available Tabs: ")
                for _, tab := range Tabs {
                    if tab.serving {
                        fmt.Printf("%d - Running a host on %d\n", tab.id, tab.port)
                    } else {
                        fmt.Printf("%d - Free\n", tab.id)
                    }
                }

                fmt.Print("\nTab Id: ")
                switchTabID := NInput()
                SwitchTab(switchTabID, false)
            case 5:
                CloseTab(CurrentTabID)
            case 6:
                currentTab := Tabs[CurrentTabID]
                // if currentTab.serving {
                //     fmt.Printf("Already serving at port: %d\n", currentTab.port)
                //     continue
                // }

                fmt.Println("Last Visited: ")
                ReadWebpagesHistory()
                for key, site := range Sites {
                    fmt.Printf(
                        "%d - %s | Last Updated: %s\n",
                        key,
                        site.Name,
                        site.UpdatedAt.Format("2006-01-02 15:04:05"),
                    )
                }

                fmt.Print("\nSearch query: ")
                searchId := NInput()
                _, ok := Sites[searchId]
                if !ok {
                    fmt.Println("Invalid Search Suery")
                    continue
                }

                completed := make(chan bool)

                // send the query to tab channel
                currentTab.command <- Command {
                    Action: "start_server",
                    PageId:  searchId,
                    Completed: completed,
                }

                <- completed
            case 7:
                currentTab := Tabs[CurrentTabID]
                CloseHost(currentTab)
            default:
                fmt.Println("Unknown command")
        }

        fmt.Println()
    }
}

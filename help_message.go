package main

import (
    "os"
    "bufio"
    "fmt"
    "strings"
    "strconv"
)


var reader = bufio.NewReader(os.Stdin)

func SInput() string {
    line, err := reader.ReadString('\n')
    if err != nil {
        fmt.Println("Error reading input:", err)
    }
    return strings.TrimSpace(line)
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
        8: "",
        9: "Sync Webpages",
        10: "Allow Sync",
    }
    rows := 3
    cols := (len(msgs) + rows - 1) / rows

    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            i := c*rows + r
            if i < len(msgs) {
                if len(msgs[i]) == 0 {
                    fmt.Printf("%-25s", fmt.Sprintf(""))
                } else {
                    fmt.Printf("%-25s", fmt.Sprintf("%d - %s", i, msgs[i]))
                }
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
                ReadWebpagesHistory()
                currentTab := Tabs[CurrentTabID]
                // if currentTab.serving {
                //     fmt.Printf("Already serving at port: %d\n", currentTab.port)
                //     continue
                // }

                fmt.Println("Last Visited: ")
                for key, site := range Webpages {
                    fmt.Printf(
                        "%d - %s | Last Updated: %s\n",
                        key+1,
                        site.Name,
                        site.UpdatedAt.Format("2006-01-02 15:04:05"),
                    )
                }

                searchIndex := 1

                for {
                    fmt.Print("\nSearch query: ")
                    searchQuery := strings.TrimSpace(SInput())
                    parts := strings.Fields(searchQuery)

                    if len(parts) == 0 {
                        continue
                    }

                    if strings.HasPrefix(parts[0], "-y") {
                        if len(parts) > 1 {
                            n, err := strconv.Atoi(parts[1])
                            if err != nil {
                                fmt.Println("invalid number")
                                continue
                            }

                            searchIndex = n
                        }
                        break
                    }else if strings.HasPrefix(parts[0], "-n") {
                        searchIndex = 0
                        break
                    }

                    ReadWebpagesFolder(searchQuery)
                    for key, site := range Webpages {
                        fmt.Printf(
                            "%d - %s | Last Updated: %s\n",
                            key,
                            site.Name,
                            site.UpdatedAt.Format("2006-01-02 15:04:05"),
                        )
                    }
                }

                if searchIndex != 0 {
                    completed := make(chan bool)

                    // send the query to tab channel
                    currentTab.command <- Command {
                        Action: "start_server",
                        PageIndex:  searchIndex-1,
                        Completed: completed,
                    }

                    <- completed
                }
            case 7:
                currentTab := Tabs[CurrentTabID]
                CloseHost(currentTab)
            case 9:
                fmt.Println("Syncing...")
                StartBeacon()
            default:
                fmt.Println("Unknown command")
        }

        fmt.Println()
    }
}

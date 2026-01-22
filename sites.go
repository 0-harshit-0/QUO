package main

import (
    "os"
    "path/filepath"
    "time"
    "encoding/json"
    "log"
    // "fmt"
)


type SiteFolder struct {
    ID        int
    Name      string
    Path      string
    UpdatedAt time.Time
}

var Sites = make(map[int]*SiteFolder)



// save and load folder_name, and then use that names to load them in Sites using the ReactWebpagesFolder
func ReadWebpagesHistory() {
    rootWebpagesFolder := "webpages" // root webpages folder

    jsonData, err := os.ReadFile(CacheDir+"/history.json")
    if err != nil {
        log.Printf("failed to read json file: %v\n", err)
        return
    }

    // Define a slice to hold the data
    var history []string

    // Unmarshal the byte slice into the Response struct
    if err := json.Unmarshal(jsonData, &history); err != nil {
        log.Printf("failed to unmarshal json data: %v\n", err)
        return
    }

    id := 1
    for _, entry := range history {
        path := filepath.Join(rootWebpagesFolder, entry)
        // fmt.Println(entry, path)
        Sites[id] = &SiteFolder{
            ID:        id,
            Name:      entry,
            Path:      path,
            UpdatedAt: time.Now(), // needs to be updated and fixed
        }
        id++
    }
}


func ReadWebpagesFolder(start int, limit int) {
    rootWebpagesFolder := "webpages" // root webpages folder
    
    entries, err := os.ReadDir(rootWebpagesFolder)
    if err != nil {
        panic(err)
    }
    // unregistered port: 49152 - 65535

    if start > len(entries) {
        start = 0
    }
    if limit > len(entries) {
        limit = len(entries)
    }
    end := start + limit

    id := 1
    for _, entry := range entries[start:end] {
        if entry.IsDir() {
            path := filepath.Join(rootWebpagesFolder, entry.Name())

            info, err := os.Stat(path)
            if err != nil {
                continue
            }

            Sites[id] = &SiteFolder{
                ID:        id,
                Name:      entry.Name(),
                Path:      path,
                UpdatedAt: info.ModTime(),
            }

            id++
        }
    }
}

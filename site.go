package main

import (
    "os"
    "path/filepath"
    "time"
)


type SiteFolder struct {
    ID        int
    Name      string
    Path      string
    UpdatedAt time.Time
}

var Sites = make(map[int]*SiteFolder)

func ReadWebpagesFolder() {
    rootWebpagesFolder := "webpages" // root webpages folder
    id := 1
    
    entries, err := os.ReadDir(rootWebpagesFolder)
    if err != nil {
        panic(err)
    }
    // unregistered port: 49152 - 65535

    for _, entry := range entries {
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

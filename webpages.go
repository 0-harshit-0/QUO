package main

import (
    "os"
    "path/filepath"
    "time"
    "encoding/json"
    "log"
    "strings"
    // "fmt"
)


type WebpageFolder struct {
    ID        int
    Name      string
    Path      string
    UpdatedAt time.Time
}

var Webpages []*WebpageFolder



func ReadJson[T any](path string) (T, error) {
    var data T

    jsonData, err := os.ReadFile(path)
    if err != nil {
        log.Printf("failed to read json file: %v\n", err)
        return data, err
    }

    // Unmarshal the byte slice into the Response struct
    if err := json.Unmarshal(jsonData, &data); err != nil {
        log.Printf("failed to unmarshal json data: %v\n", err)
        return data, err
    }

    return data, nil
}


// save and load folder_name, and then use that names to load them in Webpages
func ReadWebpagesHistory() {
    // Define a slice to hold the data
    history, err := ReadJson[[]string](CacheDir+"/history.json")
    if err != nil {
        return
    }

    id := 1
    Webpages = nil
    for _, entry := range history {
        path := filepath.Join(RootWebpagesFolder, entry)
        // fmt.Println(entry, path)
        Webpages = append(Webpages, &WebpageFolder{
            ID:        id,
            Name:      entry,
            Path:      path,
            UpdatedAt: time.Now(), // needs to be updated and fixed
        })
        id++
    }
}


func ReadWebpagesFolder(search string) {
    //start int, limit int
    entries, err := os.ReadDir(RootWebpagesFolder)
    if err != nil {
        panic(err)
    }


    Webpages = nil
    id := 1
    for _, entry := range entries {
        if entry.IsDir() {
            // search filter
            if search != "" && !strings.HasPrefix(entry.Name(), search) {
                continue
            }

            path := filepath.Join(RootWebpagesFolder, entry.Name())

            info, err := os.Stat(path)
            if err != nil {
                continue
            }

            Webpages = append(Webpages, &WebpageFolder{
                ID:        id,
                Name:      entry.Name(),
                Path:      path,
                UpdatedAt: info.ModTime(),
            })
            id++

            if len(Webpages) > 9 {
                break
            }
        }
    }
}


func UpdateHistory(name string) (error) {
    path := CacheDir+"/history.json"

    history, err := ReadJson[[]string](path)
    if err != nil {
        return err
    }

    history = append([]string{name}, history...)
    // trim to max 10
    if len(history) > 10 {
        history = history[:10]
    }

    // write back to file
    out, err := json.MarshalIndent(history, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, out, 0644)
}
package main

import (
    "os"
    "path/filepath"
    "encoding/json"
    "log"
)


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
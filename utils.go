package main

import (
    "os"
    "encoding/json"
    "fmt"
)


func ReadJson[T any](path string) (T, error) {
    var data T

    jsonData, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("failed to read json file: %v\n", err)
        return data, err
    }

    // Unmarshal the byte slice into the Response struct
    if err := json.Unmarshal(jsonData, &data); err != nil {
        fmt.Printf("failed to unmarshal json data: %v\n", err)
        return data, err
    }

    return data, nil
}

func PrintInRows(rows int, msgs []string) {
    cols := (len(msgs) + rows - 1) / rows
    for r := 0; r < rows; r++ {
        if r%3 == 0 {
            fmt.Println()
        }
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
}
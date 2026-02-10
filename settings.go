package main

import (
    "fmt"
    "reflect"
)


type settingsJson struct {
    Discoverable         bool `json:"discoverable"`
    Receiver             bool `json:"receiver"`
}

func ShowSettings() {
    settings, err := ReadJson[settingsJson](ConfigDir+"/settings.json")
    if err != nil {
        return
    }

    v := reflect.ValueOf(settings)

    fmt.Print(v)
    // PrintInRows(3, settings)
}
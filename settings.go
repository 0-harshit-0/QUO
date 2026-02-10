package main

import (
    "fmt"
    "strings"
    "reflect"
)


type settingsJson struct {
    Receiver             bool `json:"receiver"`
    Discoverable         bool `json:"discoverable"`
    AllowSync            bool `json:"allow_sync"`
}

func ShowSettings() {
    settings, err := ReadJson[settingsJson](ConfigDir+"/settings.json")
    if err != nil {
        return
    }

    v := reflect.ValueOf(settings)
    t := reflect.TypeOf(settings)

    for i := 0; i < v.NumField(); i++ {
        // key := t.Field(i).Tag.Get("json")
        key := strings.ToLower(t.Field(i).Name)
        value := v.Field(i).Interface()

        fmt.Printf("%d - %s --- %v\n", i+1, key, value)
    }

    // fmt.Print(v)
    // fmt.Print(t)
    // PrintInRows(3, settings)
}
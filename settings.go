package main

import (
    "os"
    "fmt"
    "strings"
    "reflect"
    "encoding/json"
)


type settingsJson struct {
    Receiver             bool `json:"receiver"`
    AllowSync            bool `json:"allow_sync"`
}
var Settings settingsJson


func LoadSettings() {
    Logger.Info("Loading Settings Config File")

    var err error
    Settings, err = ReadJson[settingsJson](ConfigDir+"/settings.json")
    if err != nil {
        return
    }
}

func ListSettings() {
    // add a restart flag, this way if user has not restarted we can safely ignore calling other functions or restart automatically
    fmt.Println("Restart the browser after updating the settings")

    v := reflect.ValueOf(Settings)
    t := reflect.TypeOf(Settings)

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

func UpdateSetting(id int) error {
    switch id {
        case 1:
            Settings.Receiver = !Settings.Receiver
        case 2:
            Settings.AllowSync = !Settings.AllowSync
        default:
            fmt.Println("\nInvalid setting\n")
    }

    Logger.Info("Updating Settings Config File")

    data, err := json.MarshalIndent(Settings, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(ConfigDir+"/settings.json", data, 0644)
}

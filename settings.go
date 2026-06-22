package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type settingsJson struct {
	Sync         bool `json:"sync"`
	SendNodes    bool `json:"send_nodes"`
	ReceiveNodes bool `json:"receive_nodes"`
	UseTempIP6   bool `json:"use_temp_ip6"`
}

// user should not update the JSON directly. Even if they do, only refresh it on broser start
var settingsFileName = "/settings.json"
var Settings settingsJson

func LoadSettings() {
	Logger.Info("Loading Settings Config File")

	var err error
	Settings, err = ReadJson[settingsJson](Configs.JsonConfigsDir + settingsFileName)
	if err != nil {
		Logger.Error("Cannot load settings file", "error", err)
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
		key := t.Field(i).Name
		value := v.Field(i).Interface()

		fmt.Printf("%d - %s --- %v\n", i+1, key, value)
	}

	// fmt.Print(v)
	// fmt.Print(t)
	// PrintInRows(3, settings)
}

func UpdateSetting(id int) error {
	Logger.Info("Updating Settings Config File")

	switch id {
	case 1:
		Settings.Sync = !Settings.Sync
	case 2:
		Settings.SendNodes = !Settings.SendNodes
	case 3:
		Settings.ReceiveNodes = !Settings.ReceiveNodes
	case 4:
		Settings.UseTempIP6 = !Settings.UseTempIP6
	default:
		fmt.Println("\nInvalid setting")
	}

	data, err := json.MarshalIndent(Settings, "", "  ")
	if err != nil {
		Logger.Error("Settings did not updated", "error", err)
		return err
	}

	// Logger.Info("Saved the Updated Settings Config File")
	return os.WriteFile(Configs.JsonConfigsDir+settingsFileName, data, 0644)
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	// "fmt"
)

// each webpage has a hash: randomly 3 files, could be the same file, randomly 3 lines from each, along with the author name etc, and create a hash
type webpageFolder struct {
	ID        int
	Name      string
	Path      string
	UpdatedAt time.Time
}

// user can update the JSON manually. No need to restart the browser
var RootWebpagesFolder string = "webpages" // root webpages folder
var Webpages = make([]*webpageFolder, 0, 9)

// NOT USED ANYWHERE FOR NOW
func LoadWebpagesFolder() ([]string, error) {
	Logger.Info("Loading Webpages Folder")

	var webpagesName []string

	entries, err := os.ReadDir(RootWebpagesFolder)
	if err != nil {
		Logger.Error("Cannot read webpages folder", "error", err)
		return nil, err
	}

	// Iterate over the entries
	for _, file := range entries {
		if file.IsDir() {
			webpagesName = append(webpagesName, file.Name())
		}
	}

	return webpagesName, nil
}

func SearchWebpagesFolder(search string) {
	Logger.Info("Searching Webpages Folder")

	//start int, limit int
	entries, err := os.ReadDir(RootWebpagesFolder)
	if err != nil {
		Logger.Error("Cannot read webpages folder", "error", err)
		panic(err)
	}

	id := 1
	Webpages = Webpages[:0]
	for _, entry := range entries {
		if entry.IsDir() {
			// search filter
			if search != "" && !strings.HasPrefix(entry.Name(), search) {
				continue
			}

			path := filepath.Join(RootWebpagesFolder, entry.Name())

			info, err := os.Stat(path)
			if err != nil {
				Logger.Error("No stat", "error", err)
				continue
			}

			Webpages = append(Webpages, &webpageFolder{
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

// save and load folder_name, and then use that names to load them in Webpages
func ReadWebpagesHistory() {
	Logger.Info("Loading Webpages History File")

	// Define a slice to hold the data
	history, err := ReadJson[[]string](CacheDir + "/history.json")
	if err != nil {
		Logger.Error("Error Loading Webpages History File", "error", err)
		return
	}

	id := 1
	Webpages = Webpages[:0]
	for _, entry := range history {
		path := filepath.Join(RootWebpagesFolder, entry)
		// fmt.Println(entry, path)
		Webpages = append(Webpages, &webpageFolder{
			ID:        id,
			Name:      entry,
			Path:      path,
			UpdatedAt: time.Now(), // needs to be updated and fixed
		})
		id++
	}
}

func UpdateHistory(name string) error {
	Logger.Info("Updating webpages history")

	path := CacheDir + "/history.json"

	history, err := ReadJson[[]string](path)
	if err != nil {
		Logger.Error("Cannot webpages history", "error", err)
		return err
	}

	history = append([]string{name}, history...)
	// trim to max 9
	if len(history) > 9 {
		history = history[:9]
	}

	// write back to file
	out, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		Logger.Error("Cannot update webpages history", "error", err)
		return err
	}

	// Logger.Info("Update webpages history")
	return os.WriteFile(path, out, 0644)
}

func ListWebpages() {
	for key, site := range Webpages {
		fmt.Printf(
			"%d - %s | Last Updated: %s\n",
			key+1,
			site.Name,
			site.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
	}
}

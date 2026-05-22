package main

import (
	"encoding/json"
	"os"
)

type nodeJson struct {
	Addr         string `json:"addr"`
	Port         int    `json:"port"`
	CheckedCount int    `json:"checked_count"`
}

// user can update the JSON manually. No need to restart the browser
var AllNodes []nodeJson

func LoadNodes() {
	Logger.Info("Loading Nodes Config File")

	AllNodes = AllNodes[:0]

	nodes, err := ReadJson[[]nodeJson](ConfigDir + "/nodes.json")
	if err != nil {
		Logger.Error("Error loading nodes", "error", err)
		return
	}

	for _, n := range nodes {
		if n.CheckedCount < 6 {
			AllNodes = append(AllNodes, n)
		}
	}

	// return AllNodes
}

func UpdateNodes(ip string, port int) {
	Logger.Info("Updating node")

	newNode := nodeJson{
		Addr:         ip,
		Port:         port,
		CheckedCount: 0,
	}

	AllNodes = append(AllNodes, newNode)
}

func SaveNodes() {
	Logger.Info("Saving nodes")
	path := ConfigDir + "/nodes.json"

	// write back to file
	out, err := json.MarshalIndent(AllNodes, "", "  ")
	if err != nil {
		Logger.Error("Error saving nodes", "error", err)
		return
	}

	os.WriteFile(path, out, 0644)
}

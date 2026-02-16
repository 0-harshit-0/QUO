package main


import (
    "os"
    "encoding/json"
)


type nodeJson struct {
    Addr         string `json:"addr"`
    Port         int    `json:"port"`
    CheckedCount int    `json:"checked_count"`
}


var AllNodes []nodeJson


func LoadNodes(maxCount int) {
    Logger.Info("Loading Nodes Config File")

	nodes, err := ReadJson[[]nodeJson](ConfigDir+"/nodes.json")
    if err != nil {
    	Logger.Error("Error loading nodes", err)
        return
    }

    for _, n := range nodes {
    	if n.CheckedCount < maxCount {
    		AllNodes = append(AllNodes, n)
    	}
	}

	// return AllNodes
}

func UpdateNodes(ip string, port int) {
    newNode := nodeJson{
        Addr: ip,
        Port: port,
        CheckedCount: 0,
    }

    AllNodes = append(AllNodes, newNode)
}

func SaveNodes() {
    path := ConfigDir+"/nodes.json"

    // write back to file
    out, err := json.MarshalIndent(AllNodes, "", "  ")
    if err != nil {
    	Logger.Error("Error saving nodes", err)
        return
    }

    os.WriteFile(path, out, 0644)
}

package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	structViewer "struct-viewer"
)

const PREFIX = "TYK_"

func main() {
	anyConfig := structViewer.ExampleConfig{
		ExportedField: "val",
		Multiple:      []structViewer.Single{"one", "two"},
		InnerObject: structViewer.Obj{
			Temp:    32,
			Timeout: 2 * time.Second,
			Map: structViewer.KV{
				Key:   "key",
				Value: "value",
			},
		},
	}

	viewer := structViewer.New(&structViewer.Config{Object: anyConfig}, PREFIX)
	if err := viewer.ParseComments(); err != nil {
		log.Fatalf("failed to parse comments, err: %v", err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(viewer.Envs()); err != nil {
		log.Fatalf("failed to encode environments, err: %v", err)
	}
}

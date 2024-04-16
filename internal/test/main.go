package main

import (
	"log"
	"net/http"

	structviewer "github.com/TykTechnologies/structviewer"
)

type Config struct {
	// ListenPort represents the port to listen on.
	ListenPort int `json:"listen_port"`
	// Debug represents the debug mode.
	Debug bool `json:"debug"`
	// LogFile represents the log file path.
	LogFile string `json:"log_file"`
}

func main() {
	config := &structviewer.Config{
		Object: &Config{
			ListenPort: 8080,
			Debug:      true,
			LogFile:    "/var/log/app.log",
		},
		Path:          "./main.go",
		ParseComments: true,
	}

	// prefix is added to each env var
	v, err := structviewer.New(config, "APP_")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/config", v.JSONHandler)
	http.HandleFunc("/envs", v.EnvsHandler)
	log.Println("Running server on port :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

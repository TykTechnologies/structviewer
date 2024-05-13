package main

import (
	"log"
	"net/http"

	"github.com/TykTechnologies/structviewer"
)

type complexType struct {
	Data struct {
		Object1 int  `json:"object_1,omitempty"`
		Object2 bool `json:"object_2,omitempty"`
	} `json:"data"`
	Metadata map[string]struct {
		ID    int    `json:"id,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"metadata,omitempty"`
	Random map[int]string `json:"random,omitempty"`
}

var complexStruct = complexType{
	Data: struct {
		Object1 int  `json:"object_1,omitempty"`
		Object2 bool `json:"object_2,omitempty"`
	}{
		Object1: 1,
		Object2: true,
	},
	Metadata: map[string]struct {
		ID    int    `json:"id,omitempty"`
		Value string `json:"value,omitempty"`
	}{
		"key_99": {ID: 99, Value: "key99"},
	},
	Random: map[int]string{
		1: "one",
		2: "two",
	},
}

func main() {
	config := &structviewer.Config{
		Object:        complexStruct,
		Path:          "./main.go",
		ParseComments: true,
	}

	// prefix is added to each env var
	v, err := structviewer.New(config, "APP_")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/config", v.ConfigHandler)
	http.HandleFunc("/detailed-config", v.DetailedConfigHandler)
	http.HandleFunc("/envs", v.EnvsHandler)

	log.Println("Running server on port :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

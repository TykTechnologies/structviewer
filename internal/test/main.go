package main

import (
	"log"
	"net/http"

	structviewer "github.com/TykTechnologies/structviewer"
)

type complexType struct {
	// Name represents a name.
	Name string `json:"name,omitempty"`
	// Data represents a data structure.
	Data struct {
		// Object1 represents an integer.
		Object1 int `json:"object_1,omitempty"`
		// Object2 represents a boolean.
		Object2 bool `json:"object_2,omitempty"`
	} `json:"data"`
	// Metadata represents a metadata map.
	Metadata map[string]struct {
		// ID represents an integer.
		ID int `json:"id,omitempty"`
		// Value represents a string.
		Value string `json:"value,omitempty"`
	} `json:"metadata,omitempty"`
	// OmittedValue represents an omitted value.
	OmittedValue string `json:"omitted_value,omitempty"`
}

var complexStruct = complexType{
	Name: "name_value",
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
}

func main() {
	config := &structviewer.Config{
		Object:        &complexStruct,
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

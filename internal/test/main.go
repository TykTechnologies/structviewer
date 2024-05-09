package main

import (
	"log"
	"net/http"

	"github.com/TykTechnologies/structviewer"
)

// InnerStructType represents an inner struct.
type InnerStructType struct {
	// DummyAddr represents an address.
	DummyAddr string `json:"dummy_addr"`
}

// StructType represents a struct type.
type StructType struct {
	// Enable represents status.
	Enable bool `json:"enable"`
	// Inner is an inner struct.
	Inner InnerStructType `json:"inner"`
}

type testStruct struct {
	// Exported represents a sample exported field.
	Exported    string `json:"exported"`
	notExported bool   //lint:ignore U1000 Unused field is used for testing purposes.

	// StrField is a struct field.
	StrField struct {
		// Test is a field of struct type.
		Test  string `json:"test"`
		Other struct {
			// OtherTest represents a field of sub-struct.
			OtherTest   bool `json:"other_test"`
			nonEmbedded string
		} `json:"other"`
	} `json:"str_field"`

	// ST is another struct type.
	ST StructType `json:"st"`

	// JSONExported includes a JSON tag.
	JSONExported int `json:"json_exported" structviewer:"obfuscate"`
}

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

package main

import (
	"log"
	"net/http"

	"github.com/TykTechnologies/structviewer"
)

type InnerStructType struct {
	// DummyAddr represents an address.
	DummyAddr string `json:"dummy_addr"`
}

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

	// JsonExported includes a JSON tag.
	JsonExported int `json:"json_exported"`
}

func main() {
	config := &structviewer.Config{
		Object:        &testStruct{},
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

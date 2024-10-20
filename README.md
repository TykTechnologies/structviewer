# struct_viewer

The struct_viewer package is a Go library designed to simplify the management and visualization of application configurations using Go structs and environment variables. It allows you to define your configuration as a struct, populate it with values from environment variables, and expose both the configuration and its corresponding environment variables via HTTP endpoints for easy inspection and debugging.

## Installation

To install the struct_viewer package, use the go get command:

```bash
go get -u github.com/TykTechnologies/struct-viewer
```

## Usage

```go
package main

import (
	"log"
	"net/http"

	"github.com/TykTechnologies/structviewer"
)

type Config struct {
	ListenPort int    `json:"listen_port"`
	Debug      bool   `json:"debug"`
	LogFile    string `json:"log_file"`
}

func main() {
	config := &struct_viewer.Config{
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
```

This will expose two endpoints:

- `/config`: returns a json representation of the config struct
- `/envs`: returns a json representation of the environment variables of the config struct

You can pass query parameters `field` or `env` to the endpoints to retrieve specific config fields or specific environment variables.

## Limitations

- Only exported fields in struct are parsed
- Only struct fields can have comments in them
- Single file parsing
- No obfuscation

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request on GitHub.

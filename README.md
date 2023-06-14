# struct_viewer

The struct_viewer package is a Go library that provides functionality for managing application configurations using Go struct types and environment variables. It allows you to define your configuration as a struct and easily populate it with values from environment variables

## Installation

To install the struct_viewer package, use the go get command:

```bash
go get -u github.com/TykTechnologies/struct_viewer
```

## Usage
```go
package main

import (
    "fmt"
    "net/http"

    "github.com/yourusername/struct-viewer"
)

type Config struct {
    ListenPort int    `env:"listen_port"`
    Debug      bool   `env:"debug"`
    LogFile    string `env:"log_file"`
}

func main() {
    config := &struct_viewer.Config{
        Object: &Config{
            ListenPort: 8080,
            Debug:      true,
            LogFile:    "/var/log/app.log",
        },
        Path:          "./config.go",
        ParseComments: true,
    }

    // prefix is added to each env var
    v, err := struct_viewer.New(config, "APP_")
    if err != nil {
        panic(err)
    }

    http.HandleFunc("/config", v.JSONHandler)
    http.HandleFunc("/envs", v.EnvsHandler)
    http.ListenAndServe(":8080", nil)
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


##  Contributing
Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request on GitHub.
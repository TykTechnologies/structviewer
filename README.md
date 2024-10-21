# struct_viewer

The struct_viewer package is a Go library designed to simplify the management and visualization of application configurations using Go structs and environment variables. It allows you to define your configuration as a struct, populate it with values from environment variables, and expose both the configuration and its corresponding environment variables via HTTP endpoints for easy inspection and debugging.

## Features

- Parse struct-based configurations.
- Expose configuration fields and environment variables as JSON.
- Simple HTTP handlers to expose config and envs.
- Secure: Automatically obfuscate sensitive data.

## Installation

To install the *struct_viewer* package, use the `go get` command:

```bash
go get -u github.com/TykTechnologies/struct-viewer
```

## Usage

Below is a simple example of how to use *struct_viewer* in your application:

```go
package main

import (
	"log"
	"net/http"

	"github.com/TykTechnologies/structviewer"
)

type Config struct {
	
	Field1       string `json:"field1"`
	Field2       int    `json:"field2"`

	// This field can be obfuscated with the structviewer tag
	Field3secret string `json:"field3_secret" structviewer:"obfuscate"` 
}

func main() {

	  // Define configuration structure
  	config := &structviewer.Config{
               Object: &YourConfigStruct{
   		       Field1: "value1",
  		       Field2: 123,
   	       },

        }

	// prefix is added to each env var
	v, err := structviewer.New(config, "PREFIX")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/config", v.JSONHandler)
	http.HandleFunc("/envs", v.EnvsHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Println("Running server on port :8080")
}
```

This will expose two endpoints:

- `/config`: returns a json representation of the config struct.
  - Example: `curl -s 'http://localhost:8080/config' | jq .`
- `/envs`: returns a json representation of the environment variables of the config struct.
   You can pass query parameters `env` to the endpoints to retrieve specific config fields or specific environment variables.
   - Example:
     - `curl -s 'http://localhost:8080/env' | jq .`
     - `curl -s 'http://localhost:8080/env?env=PREFIX_FIELD1' | jq .`

## HTTP Handlers
`/config`: Exposes the entire config object.
`/detailed-config`: Exposes detailed configuration fields with descriptions.
`/envs`: Exposes environment variables mapped from the config object.


## Error Handling
The library provides several error types:

`ErrNilConfig`: Returned when a nil configuration struct is provided.
`ErrEmptyStruct`: Returned when an empty struct is provided.
`ErrInvalidObjectType`: Returned when the object is not of struct type.
Contributing
We welcome contributions! Please see our contribution guidelines for details.

By incorporating these elements, your README will be more accessible and informative for users and contributors alike.


## Limitations

- Only exported fields in go struct are parsed
- Only struct fields can have comments in them
- Single file parsing
- No obfuscation

## Contributing

Contributing
We welcome contributions! Please see our [contribution guidelines](CONTRIBUTING.md) for details. If you find any issues or have suggestions for improvement, please open an issue or submit a pull request on GitHub.

package struct_viewer

import (
	"bytes"
	"time"
)

type KV struct {
	// Key represents a key for your KV.
	Key string `json:"key"`
	// Value represents a value of specific Key.
	Value interface{} `json:"value"`
}

type Obj struct {
	// unexported field for Obj.
	unexported string

	// Temp value for Obj.
	Temp   int64        `json:"temp"`
	Buffer bytes.Buffer `json:"buffer"`
	// Timeout represents timeout for Obj.
	Timeout time.Duration `json:"timeout"`
	// Map represents Obj's KV.
	Map KV `json:"map"`
}

type Single string

type ExampleConfig struct {
	// ExportedField represents exported struct field.
	ExportedField string `json:"exported_field,omitempty"`

	// Multiple represents custom object type.
	Multiple []Single `json:"multiple,omitempty"`

	// InnerObject includes a field that includes struct objects.
	InnerObject Obj `json:"inner_object"`
}

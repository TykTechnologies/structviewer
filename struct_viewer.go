package struct_viewer

import (
	"errors"
	"go/ast"
)

// Viewer is the pkg control structure where the prefix and env vars are stored.
type Viewer struct {
	config       interface{}
	prefix       string
	confFilePath string

	envs []*EnvVar
	file *ast.File
}

var (
	ErrNilConfig   = errors.New("invalid Config structure provided")
	ErrEmptyStruct = errors.New("empty Struct in configuration")
)

// Config represents configuration structure.
type Config struct {
	// Object represents an object that is going to be parsed.
	Object interface{}

	// Path is the file path of the Object. Needed for comment parser.
	// Default value is "./config.go".
	Path string
}

// New receives a configuration structure and a prefix and returns a Viewer struct to manipulate this library.
func New(config *Config, prefix string) (*Viewer, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if config.Object == nil {
		return nil, ErrEmptyStruct
	}

	if config.Path == "" {
		config.Path = "./config.go"
	}

	cfg := Viewer{config: config.Object, prefix: prefix, confFilePath: config.Path}
	cfg.Start()

	return &cfg, nil
}

// Start starts the Viewer control struct, parsing the environment variables
func (v *Viewer) Start() {
	v.envs = parseEnvs(v.config, v.prefix)
}

func (v *Viewer) ParseComments() error {
	return v.parseComments()
}

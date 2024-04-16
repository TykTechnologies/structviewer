package struct_viewer

import (
	"errors"
	"go/ast"

	"github.com/fatih/structs"
)

// Viewer is the pkg control structure where the prefix and env vars are stored.
type Viewer struct {
	config       interface{}
	prefix       string
	confFilePath string

	envs      []*EnvVar
	configMap map[string]interface{}
	file      *ast.File
}

var (
	ErrNilConfig         = errors.New("invalid Config structure provided")
	ErrEmptyStruct       = errors.New("empty Struct in configuration")
	ErrInvalidObjectType = errors.New("invalid object type")
)

// Config represents configuration structure.
type Config struct {
	// Object represents an object that is going to be parsed.
	Object interface{}

	// ParseComments decides parsing comments of given Object or not. If it is set to false,
	// the comment parser skips parsing comments of given Object.
	ParseComments bool

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

	if !structs.IsStruct(config.Object) {
		return nil, ErrInvalidObjectType
	}

	if config.Path == "" {
		config.Path = "./config.go"
	}

	cfg := Viewer{config: config.Object, prefix: prefix, confFilePath: config.Path}
	err := cfg.start(config.ParseComments)

	return &cfg, err
}

// Start starts the Viewer control struct, parsing the environment variables
func (v *Viewer) start(parseComments bool) error {
	v.envs = parseEnvs(v.config, v.prefix)
	if parseComments {
		v.parseComments()
	}

	v.configMap = v.parseConfig()

	return nil
}

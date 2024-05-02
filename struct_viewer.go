package structviewer

import (
	"errors"
	"go/ast"
	"reflect"

	"github.com/fatih/structs"
)

// Viewer is the pkg control structure where the prefix and env vars are stored.
type Viewer struct {
	// config is the configuration structure.
	config interface{}
	// prefix is the prefix of the environment variables.
	prefix string
	// confFilePath is the file path of the configuration structure.
	confFilePath string

	// envs is the slice of environment variables.
	// It is used to expose the environment variables as JSON in EnvsHandler.
	envs []*EnvVar
	// configMap is the map representation of the configuration structure.
	// It is used to expose the configuration structure as JSON in JSONHandler.
	configMap map[string]*EnvVar
	// file is the ast.File of the configuration structure.
	file *ast.File
	// obfuscatedTags is the list of JSON tags that should be obfuscated.
	obfuscatedTags []string
}

var (
	// ErrNilConfig is returned when the Config struct is nil.
	ErrNilConfig = errors.New("invalid Config structure provided")
	// ErrEmptyStruct is returned when config.Object is nil
	ErrEmptyStruct = errors.New("empty Struct in configuration")
	// ErrInvalidObjectType is returned when config.Object is not a struct.
	ErrInvalidObjectType = errors.New("invalid object type")
	// ErrNotPointer is returned when the struct is not a pointer.
	ErrNotPointer = errors.New("config structure must be a pointer")
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

	// ObfuscatedTags is the list of JSON tags that should be obfuscated.
	// If the JSON tag of a field is in this list, the field value will be defaulted to its zero value.
	ObfuscatedTags []string
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

	// Struct must be a pointer so we can modify its values if needed.
	value := reflect.ValueOf(config.Object)
	if value.Kind() != reflect.Ptr {
		return nil, ErrNotPointer
	}

	if config.Path == "" {
		config.Path = "./config.go"
	}

	cfg := Viewer{config: config.Object, prefix: prefix, confFilePath: config.Path, obfuscatedTags: config.ObfuscatedTags}
	err := cfg.start(config.ParseComments)

	return &cfg, err
}

// Start starts the Viewer control struct, parsing the environment variables
func (v *Viewer) start(parseComments bool) error {
	var err error
	v.config, err = obfuscateTags(v.config, v.obfuscatedTags, "")
	if err != nil {
		return err
	}

	v.envs = parseEnvs(v.config, v.prefix, "", v.obfuscatedTags)
	if parseComments {
		if err = v.parseComments(); err != nil {
			return err
		}
	}

	v.configMap = parseConfig(v.envs)

	return nil
}

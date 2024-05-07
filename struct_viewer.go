package structviewer

import (
	"errors"
	"go/ast"
	"reflect"
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
}

var (
	// ErrNilConfig is returned when the Config struct is nil.
	ErrNilConfig = errors.New("invalid Config structure provided")
	// ErrEmptyStruct is returned when config.Object is nil
	ErrEmptyStruct = errors.New("empty Struct in configuration")
	// ErrInvalidObjectType is returned when config.Object is not a struct.
	ErrInvalidObjectType = errors.New("invalid object type")
)

// Config represents configuration structure.
type Config struct {
	// Object represents the config struct that is going to be parsed.
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

	objectValue := reflect.ValueOf(config.Object)
	if objectValue.Kind() != reflect.Struct &&
		!(objectValue.Kind() == reflect.Ptr && objectValue.Elem().Kind() == reflect.Struct) {
		return nil, ErrInvalidObjectType
	}

	// The struct must be a pointer to be able to modify its fields if they need to be obfuscated
	// To avoid modifying the original struct, we create a copy of it
	var objectCopy interface{}
	if objectValue.Kind() == reflect.Ptr {
		objectCopy = reflect.New(objectValue.Elem().Type()).Interface()
		reflect.ValueOf(objectCopy).Elem().Set(objectValue.Elem())
	} else {
		objectCopy = reflect.New(objectValue.Type()).Interface()
		reflect.ValueOf(objectCopy).Elem().Set(objectValue)
	}

	if config.Path == "" {
		config.Path = "./config.go"
	}

	cfg := Viewer{config: objectCopy, prefix: prefix, confFilePath: config.Path}
	err := cfg.start(config.ParseComments)

	return &cfg, err
}

// Start starts the Viewer control struct, parsing the environment variables
func (v *Viewer) start(parseComments bool) error {
	var err error

	v.config, err = obfuscateTags(v.config)
	if err != nil {
		return err
	}

	v.envs = parseEnvs(v.config, v.prefix, "")
	if parseComments {
		if err = v.parseComments(); err != nil {
			return err
		}
	}

	v.configMap = parseConfig(v.envs)

	return nil
}

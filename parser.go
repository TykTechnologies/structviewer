package struct_viewer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/fatih/structs"
)

// ParseEnvs parse Viewer config field, generating a map[string]interface{} of prefix+key:value of each config field
func (v *Viewer) ParseEnvs() map[string]interface{} {
	envs := make(map[string]interface{})
	envVars := v.envs

	if len(envVars) == 0 {
		envVars = parseEnvs(v.config, v.prefix)
	}

	for _, value := range envVars {
		envs[v.prefix+value.key] = value.Value
	}

	return envs
}

// EnvNotation takes JSON notation of a configuration field (e.g, 'listen_port') and returns EnvVar object of the given
// notation.
func (v *Viewer) EnvNotation(jsonField string) EnvVar {
	for i := 0; i < len(v.envs); i++ {
		if jsonField == v.envs[i].ConfigField {
			return *v.envs[i]
		}
	}

	return EnvVar{}
}

// JSONNotation takes environment variable and returns EnvVars object of the given environment variable.
func (v *Viewer) JSONNotation(envVarNotation string) EnvVar {
	for i := 0; i < len(v.envs); i++ {
		if v.prefix+v.envs[i].key == envVarNotation {
			return *v.envs[i]
		}
	}

	return EnvVar{}
}

// Envs returns environment variables parsed by struct-viewer.
func (v *Viewer) Envs() []*EnvVar {
	return v.envs
}

func (v *Viewer) parseComments() error {
	// If we have already parsed comments, don't parse again.
	if v.file != nil {
		return nil
	}

	astFile, err := parser.ParseFile(token.NewFileSet(), v.confFilePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	v.file = astFile

	ast.Inspect(v.file, func(n ast.Node) bool {
		structType, ok := n.(*ast.StructType)
		if !ok {
			return true
		}

		v.parseInnerFields(structType)

		return false
	})

	return nil
}

func (v *Viewer) parseInnerFields(s *ast.StructType) {
	for _, structField := range s.Fields.List {
		comment := structField.Doc.Text()
		confField := structField.Names[0]

		envVar := v.get(confField.Name)
		if comment != "" && envVar != nil {
			envVar.Description = strings.TrimSpace(comment)
		}

		if structType, ok := structField.Type.(*ast.StructType); ok {
			v.parseInnerFields(structType)
		}
	}
}

func (v *Viewer) get(field string) *EnvVar {
	for _, e := range v.envs {
		if e.field == field {
			return e
		}
	}

	return nil
}

func parseEnvs(config interface{}, prefix string) []*EnvVar {
	var envs []*EnvVar

	s := structs.New(config)

	for _, field := range s.Fields() {
		if field.IsExported() {
			newEnv := &EnvVar{}
			newEnv.setKey(field)

			if structs.IsStruct(field.Value()) {
				envsInner := parseEnvs(field.Value(), prefix)

				for i := range envsInner {
					envsInner[i].key = newEnv.key + "_" + envsInner[i].key
					envsInner[i].ConfigField = newEnv.ConfigField + "." + envsInner[i].ConfigField
					envsInner[i].Env = prefix + envsInner[i].key
				}

				envs = append(envs, envsInner...)
			} else {
				newEnv.setValue(field)
				newEnv.Env = prefix + newEnv.key
				envs = append(envs, newEnv)
			}
		}
	}

	return envs
}

// EnvVar is a key:value string struct for environment variables representation
type EnvVar struct {
	// key represents an environment notation without prefix. It is used internally to generate environment variable
	// notations from given struct fields. For example, consider the following struct:
	//	outer_field: {
	//		inner_field: true
	//	}
	// For inner_field, the key is OUTERFIELD_INNERFIELD.
	key string `json:"-"`
	// field represents raw field names of the given struct fields.
	field string `json:"-"`

	// ConfigField represents a JSON notation of the given struct fields.
	ConfigField string `json:"config_field"`
	// Env represents an environment variable notation of the given struct fields.
	Env string `json:"env"`
	// Description represents the comment of the given struct fields.
	Description string `json:"description,omitempty"`
	// Value represents the value of the given struct fields.
	Value interface{} `json:"value"`
}

// String returns a key:value string from EnvVar
func (ev EnvVar) String() string {
	return ev.key + ":" + fmt.Sprintf("%s", ev.Value)
}

func (ev *EnvVar) setKey(field *structs.Field) {
	key := field.Name()
	jsonTag := field.Tag("json")

	if jsonTag != "" && jsonTag != "-" {
		jsonTag = strings.ReplaceAll(jsonTag, ",omitempty", "")
		key = jsonTag
	}

	key = strings.ReplaceAll(key, "_", "")
	key = strings.ToUpper(key)
	ev.key = key
	ev.ConfigField = jsonTag
	ev.field = field.Name()
}

func (ev *EnvVar) setValue(field *structs.Field) {
	if structs.IsStruct(field.Value()) {
		ev.Value = fmt.Sprintf("%+v", field.Value())
		return
	}

	ev.Value = field.Value()
}

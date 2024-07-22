package structviewer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/fatih/structs"
)

// ParseEnvs parse Viewer config field, generating a string slice of prefix+key:value of each config field
func (v *Viewer) ParseEnvs() []string {
	var envs []string
	envVars := v.configMap

	for _, envVar := range envVars {
		envs = append(envs, generateEnvStrings(envVar)...)
	}

	return envs
}

func generateEnvStrings(e *EnvVar) []string {
	var strEnvs []string

	if e.isStruct {
		typedEnv, ok := e.Value.(map[string]*EnvVar)
		if !ok {
			return []string{""}
		}

		for _, v := range typedEnv {
			strEnvs = append(strEnvs, generateEnvStrings(v)...)
		}

		return strEnvs
	}

	if e.Value == "" || e.Value == nil {
		e.Value = `''`
	}

	strEnvs = append(strEnvs, fmt.Sprintf("%v=%v", e.Env, e.Value))

	return strEnvs
}

// EnvNotation takes JSON notation of a configuration field (e.g, 'listen_port') and returns EnvVar object of the given
// notation.
func (v *Viewer) EnvNotation(jsonField string) EnvVar {
	ev := v.envNotationHelper(jsonField, v.envs)
	if ev == nil {
		ev = &EnvVar{}
	}

	return *ev
}

func (v *Viewer) envNotationHelper(jsonField string, envs []*EnvVar) *EnvVar {
	for i := 0; i < len(envs); i++ {
		if jsonField == envs[i].ConfigField {
			return envs[i]
		}

		if envs[i].isStruct {
			val, ok := envs[i].Value.(map[string]*EnvVar)
			if !ok {
				continue
			}

			temporarySlice := []*EnvVar{}
			for _, v := range val {
				temporarySlice = append(temporarySlice, v)
			}

			ev := v.envNotationHelper(jsonField, temporarySlice)
			if ev != nil {
				return ev
			}
		}
	}

	return nil
}

// JSONNotation takes environment variable and returns EnvVars object of the given environment variable.
func (v *Viewer) JSONNotation(envVarNotation string) EnvVar {
	if envVarNotation == "" {
		return EnvVar{}
	}

	ev := v.jsonNotationHelper(envVarNotation, v.envs)
	if ev == nil {
		ev = &EnvVar{}
	}

	return *ev
}

func (v *Viewer) jsonNotationHelper(envVarNotation string, envs []*EnvVar) *EnvVar {
	for i := 0; i < len(envs); i++ {
		if envs[i].Env == envVarNotation {
			return envs[i]
		}

		if envs[i].isStruct {
			val, ok := envs[i].Value.(map[string]*EnvVar)
			if !ok {
				continue
			}

			temporarySlice := []*EnvVar{}
			for _, v := range val {
				temporarySlice = append(temporarySlice, v)
			}

			ev := v.jsonNotationHelper(envVarNotation, temporarySlice)
			if ev != nil {
				return ev
			}
		}
	}

	return nil
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

func parseConfig(envs []*EnvVar) map[string]*EnvVar {
	configMap := map[string]*EnvVar{}
	for _, env := range envs {
		configMap[env.field] = env
	}

	return configMap
}

func (v *Viewer) parseInnerFields(s *ast.StructType) {
	for _, structField := range s.Fields.List {
		comment := structField.Doc.Text()
		confField := structField.Names[0]

		envVar := v.get(confField.Name, v.envs)
		if comment != "" && envVar != nil {
			envVar.Description = strings.TrimSpace(comment)
		}

		if structType, ok := structField.Type.(*ast.StructType); ok {
			v.parseInnerFields(structType)
		}
	}
}

func (v *Viewer) get(field string, envs []*EnvVar) *EnvVar {
	for _, env := range envs {
		if env.field == field {
			return env
		}

		if env.isStruct {
			val, ok := env.Value.(map[string]*EnvVar)
			if !ok {
				continue
			}

			temporarySlice := []*EnvVar{}
			for _, v := range val {
				temporarySlice = append(temporarySlice, v)
			}

			ev := v.get(field, temporarySlice)
			if ev != nil {
				return ev
			}
		}
	}

	return nil
}
func parseEnvs(config interface{}, prefix, configField string) []*EnvVar {
	var envs []*EnvVar

	s := structs.New(config)

	for _, field := range s.Fields() {
		if field.IsExported() {
			newEnv := &EnvVar{}
			newEnv.setKey(field)

			// Ensure that the configField ends with a single dot (only if it is not empty)
			if configField != "" && configField[len(configField)-1] != '.' {
				configField += "."
			}

			if structs.IsStruct(field.Value()) {
				envsInner := parseEnvs(field.Value(), prefix+newEnv.key+"_", configField+newEnv.ConfigField)
				kvEnvVar := map[string]*EnvVar{}
				for i := range envsInner {
					kvEnvVar[envsInner[i].field] = envsInner[i]
				}

				newEnv.Value = kvEnvVar
				newEnv.ConfigField = ""
				newEnv.isStruct = true

				envs = append(envs, newEnv)
			} else if reflect.ValueOf(field.Value()).Kind() == reflect.Map {
				v := reflect.ValueOf(field.Value())
				keys := v.MapKeys()

				kvEnvVar := map[string]*EnvVar{}

				for _, key := range keys {
					value := v.MapIndex(key).Interface()

					// Handle different key types by converting to a string representation
					keyStr := fmt.Sprintf("%v", key)
					mapEnv := &EnvVar{
						key:   keyStr,
						field: keyStr,
					}

					if reflect.TypeOf(value).Kind() == reflect.Struct {
						// Recursively process structs
						envsInner := parseEnvs(value, prefix+newEnv.key+"_", configField+newEnv.ConfigField)
						for i := range envsInner {
							kvEnvVar[envsInner[i].field] = envsInner[i]
						}
					} else {
						// Directly assign other map values to `mapEnv`
						mapEnv.Value = value
						envSuffix := strings.ToUpper(strings.ReplaceAll(keyStr, "_", ""))
						mapEnv.Env = prefix + newEnv.key + "_" + envSuffix
						mapEnv.ConfigField = configField + newEnv.ConfigField + "." + keyStr
						mapEnv.Obfuscated = getPointerBool(false)

						kvEnvVar[keyStr] = mapEnv
					}
				}

				newEnv.Value = kvEnvVar
				newEnv.ConfigField = ""
				newEnv.isStruct = true

				envs = append(envs, newEnv)
			} else {
				// Use the existing `setValue` function to assign the value
				newEnv.setValue(field)
				newEnv.Env = prefix + newEnv.key
				newEnv.ConfigField = configField + newEnv.ConfigField
				newEnv.Obfuscated = getPointerBool(false)

				if field.IsZero() && field.Tag(StructViewerTag) == "obfuscate" {
					newEnv.Obfuscated = getPointerBool(true)
				}

				envs = append(envs, newEnv)
			}
		}
	}

	return envs
}
func obfuscateTags(config interface{}) (interface{}, error) {
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.CanAddr() {
		return nil, fmt.Errorf("cannot address value")
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		svTag := field.Tag.Get(StructViewerTag)
		if fieldValue.Kind() == reflect.Struct {
			if strings.EqualFold(svTag, "obfuscate") {
				zeroValue := reflect.Zero(fieldValue.Type())
				fieldValue.Set(zeroValue)
			} else {
				newStruct, err := obfuscateTags(fieldValue.Addr().Interface())
				if err != nil {
					return nil, err
				}

				fieldValue.Set(reflect.ValueOf(newStruct).Elem())
			}
		} else if fieldValue.Kind() == reflect.Map {
			if strings.EqualFold(svTag, "obfuscate") {
				zeroValue := reflect.Zero(fieldValue.Type())
				fieldValue.Set(zeroValue)
			} else {
				newMap := reflect.MakeMap(fieldValue.Type())
				keys := fieldValue.MapKeys()

				for _, key := range keys {
					mapValue := fieldValue.MapIndex(key)
					if !mapValue.IsValid() {
						continue
					}

					var newValue reflect.Value

					switch mapValue.Kind() {
					case reflect.Ptr, reflect.Interface:
						elemValue := mapValue.Elem()
						if elemValue.Kind() == reflect.Struct {
							newStruct, err := obfuscateTags(elemValue.Addr().Interface())
							if err != nil {
								return nil, err
							}
							newValue = reflect.ValueOf(newStruct).Elem()
						} else {
							newValue = mapValue
						}
					case reflect.Struct:
						ptrToStruct := reflect.New(mapValue.Type())
						ptrToStruct.Elem().Set(mapValue)
						newStruct, err := obfuscateTags(ptrToStruct.Interface())
						if err != nil {
							return nil, err
						}
						newValue = reflect.ValueOf(newStruct).Elem()
					default:
						newValue = mapValue
					}

					newMap.SetMapIndex(key, newValue)
				}

				fieldValue.Set(newMap)
			}
		} else {
			if strings.EqualFold(svTag, "obfuscate") {
				if fieldValue.Kind() == reflect.String {
					fieldValue.SetString("*REDACTED*")
				} else {
					zeroValue := reflect.Zero(fieldValue.Type())
					fieldValue.Set(zeroValue)
				}
			}
		}
	}

	return config, nil
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
	// isStruct is used internally to determine whether the given struct field is a struct or not.
	isStruct bool `json:"-"`

	// ConfigField represents a JSON notation of the given struct fields.
	ConfigField string `json:"config_field,omitempty"`
	// Env represents an environment variable notation of the given struct fields.
	Env string `json:"env,omitempty"`
	// Description represents the comment of the given struct fields.
	Description string `json:"description,omitempty"`
	// Value represents the value of the given struct fields.
	Value interface{} `json:"value"`
	// Obfuscated represents whether the given struct field is obfuscated or not.
	// This is a pointer to a boolean value to distinguish between the zero value
	// and the actual value (because of the 'omitempty' tag).
	Obfuscated *bool `json:"obfuscated,omitempty"`
}

// String returns a key:value string from EnvVar
func (ev EnvVar) String() string {
	return fmt.Sprintf("%s:%s", ev.Env, ev.Value)
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

	ev.Value = fmt.Sprint(field.Value())
}

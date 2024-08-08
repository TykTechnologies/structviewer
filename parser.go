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
		if !field.IsExported() {
			continue
		}

		newEnv := createEnvVar(field)
		configField = ensureConfigFieldEndsWithDot(configField)

		switch {
		case structs.IsStruct(field.Value()):
			handleStructField(newEnv, field, prefix, configField, &envs)
		case reflect.ValueOf(field.Value()).Kind() == reflect.Map:
			handleMapField(newEnv, field, prefix, configField, &envs)
		default:
			handleSimpleField(newEnv, field, prefix, configField, &envs)
		}
	}

	return envs
}

func createEnvVar(field *structs.Field) *EnvVar {
	newEnv := &EnvVar{}
	newEnv.setKey(field)

	return newEnv
}

func ensureConfigFieldEndsWithDot(configField string) string {
	if configField != "" && configField[len(configField)-1] != '.' {
		return configField + "."
	}

	return configField
}

func handleStructField(newEnv *EnvVar, field *structs.Field, prefix, configField string, envs *[]*EnvVar) {
	envsInner := parseEnvs(field.Value(), prefix+newEnv.key+"_", configField+newEnv.ConfigField)
	kvEnvVar := makeKVEnvVar(envsInner)

	newEnv.Value = kvEnvVar
	newEnv.ConfigField = ""
	newEnv.isStruct = true

	*envs = append(*envs, newEnv)
}

func handleMapField(newEnv *EnvVar, field *structs.Field, prefix, configField string, envs *[]*EnvVar) {
	v := reflect.ValueOf(field.Value())
	keys := v.MapKeys()
	kvEnvVar := make(map[string]*EnvVar)

	for _, key := range keys {
		value := v.MapIndex(key)
		keyStr := fmt.Sprintf("%v", key)
		mapEnv := &EnvVar{key: keyStr, field: keyStr}

		if value.Kind() == reflect.Struct {
			processStructInMapForEnvs(value.Interface(), prefix, newEnv, configField, kvEnvVar)
		} else {
			processSimpleValueInMap(mapEnv, value.Interface(), prefix, newEnv, configField, kvEnvVar)
		}
	}

	newEnv.Value = kvEnvVar
	newEnv.ConfigField = ""
	newEnv.isStruct = true

	*envs = append(*envs, newEnv)
}

func processStructInMapForEnvs(value interface{},
	prefix string,
	newEnv *EnvVar,
	configField string,
	kvEnvVar map[string]*EnvVar,
) {
	envsInner := parseEnvs(value, prefix+newEnv.key+"_", configField+newEnv.ConfigField)
	for i := range envsInner {
		kvEnvVar[envsInner[i].field] = envsInner[i]
	}
}

func handleSimpleField(newEnv *EnvVar, field *structs.Field, prefix, configField string, envs *[]*EnvVar) {
	newEnv.setValue(field)
	newEnv.Env = prefix + newEnv.key
	newEnv.ConfigField = configField + newEnv.ConfigField
	newEnv.Obfuscated = getPointerBool(false)

	if field.Tag(StructViewerTag) == "obfuscate" {
		newEnv.Obfuscated = getPointerBool(true)
	}

	*envs = append(*envs, newEnv)
}

func makeKVEnvVar(envsInner []*EnvVar) map[string]*EnvVar {
	kvEnvVar := make(map[string]*EnvVar)
	for i := range envsInner {
		kvEnvVar[envsInner[i].field] = envsInner[i]
	}

	return kvEnvVar
}

func processStructInMap(mapValue reflect.Value) (reflect.Value, error) {
	ptrToStruct := reflect.New(mapValue.Type())
	ptrToStruct.Elem().Set(mapValue)

	newStruct, err := obfuscateTags(ptrToStruct.Interface())
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(newStruct).Elem(), nil
}

func processSimpleValueInMap(mapEnv *EnvVar,
	value interface{},
	prefix string,
	newEnv *EnvVar,
	configField string,
	kvEnvVar map[string]*EnvVar,
) {
	mapEnv.Value = value
	envSuffix := strings.ToUpper(strings.ReplaceAll(mapEnv.key, "_", ""))
	mapEnv.Env = prefix + newEnv.key + "_" + envSuffix
	mapEnv.ConfigField = configField + newEnv.ConfigField + "." + mapEnv.key
	mapEnv.Obfuscated = getPointerBool(false)

	kvEnvVar[mapEnv.key] = mapEnv
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

		if err := processField(fieldValue, svTag); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func processField(fieldValue reflect.Value, svTag string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		return processStructField(fieldValue, svTag)
	case reflect.Map:
		return processMapField(fieldValue, svTag)
	default:
		processSimpleField(fieldValue, svTag)
	}

	return nil
}

func processStructField(fieldValue reflect.Value, svTag string) error {
	if strings.EqualFold(svTag, "obfuscate") {
		zeroValue := reflect.Zero(fieldValue.Type())
		fieldValue.Set(zeroValue)

		return nil
	}

	newStruct, err := obfuscateTags(fieldValue.Addr().Interface())
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(newStruct).Elem())

	return nil
}

func processMapField(fieldValue reflect.Value, svTag string) error {
	if strings.EqualFold(svTag, "obfuscate") {
		zeroValue := reflect.Zero(fieldValue.Type())
		fieldValue.Set(zeroValue)

		return nil
	}

	newMap := reflect.MakeMap(fieldValue.Type())
	keys := fieldValue.MapKeys()

	for _, key := range keys {
		mapValue := fieldValue.MapIndex(key)
		if !mapValue.IsValid() {
			continue
		}

		newValue, err := processMapValue(mapValue)
		if err != nil {
			return err
		}

		newMap.SetMapIndex(key, newValue)
	}

	fieldValue.Set(newMap)

	return nil
}

func processMapValue(mapValue reflect.Value) (reflect.Value, error) {
	switch mapValue.Kind() {
	case reflect.Ptr, reflect.Interface:
		return processPointerOrInterface(mapValue)
	case reflect.Struct:
		return processStructInMap(mapValue)
	default:
		return mapValue, nil
	}
}

func processPointerOrInterface(mapValue reflect.Value) (reflect.Value, error) {
	elemValue := mapValue.Elem()
	if elemValue.Kind() == reflect.Struct {
		newStruct, err := obfuscateTags(elemValue.Addr().Interface())
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(newStruct).Elem(), nil
	}

	return mapValue, nil
}

func processSimpleField(fieldValue reflect.Value, svTag string) {
	if strings.EqualFold(svTag, "obfuscate") {
		if fieldValue.Kind() == reflect.String {
			if fieldValue.String() != "" {
				fieldValue.SetString("*REDACTED*")
			}
		} else {
			zeroValue := reflect.Zero(fieldValue.Type())
			if !reflect.DeepEqual(fieldValue.Interface(), zeroValue.Interface()) {
				fieldValue.Set(zeroValue)
			}
		}
	}
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

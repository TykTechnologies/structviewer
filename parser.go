package struct_viewer

import (
	"fmt"
	"strings"

	"github.com/fatih/structs"
)

//ParseEnvs parse Viewer config field, generating a string slice of prefix+key:value of each config field
func (h *Viewer) ParseEnvs() []string {
	var envs []string
	envVars := h.envs

	if len(envs) == 0 {
		envVars = parseEnvs(h.config)
	}

	for i := range envVars {
		env := envVars[i]
		envs = append(envs, h.prefix+env.String())
	}

	return envs
}

func parseEnvs(config interface{}) []EnvVars {
	var envs []EnvVars

	s := structs.New(config)

	for _, field := range s.Fields() {
		if field.IsExported() {
			newEnv := EnvVars{}
			newEnv.setKey(field)

			if structs.IsStruct(field.Value()) {
				envsInner := parseEnvs(field.Value())

				for i := range envsInner {
					envsInner[i].Key = newEnv.Key + "_" + envsInner[i].Key
				}

				envs = append(envs, envsInner...)
			} else {
				newEnv.setValue(field)
				envs = append(envs, newEnv)
			}
		}
	}

	return envs
}

//EnvVars is a key:value string struct for environment variables representation
type EnvVars struct {
	Key   string
	Value string
}

//String returns a key:value string from EnvVars
func (ev EnvVars) String() string {
	return ev.Key + ":" + ev.Value
}

func (ev *EnvVars) setKey(field *structs.Field) {
	key := field.Name()
	jsonTag := field.Tag("json")

	if jsonTag != "" && jsonTag != "-" {
		jsonTag = strings.ReplaceAll(jsonTag, ",omitempty", "")
		key = jsonTag
	}

	key = strings.ReplaceAll(key, "_", "")
	key = strings.ToUpper(key)
	ev.Key = key
}

func (ev *EnvVars) setValue(field *structs.Field) {
	ev.Value = fmt.Sprint(field.Value())
}

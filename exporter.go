package structviewer

import (
	"encoding/json"
	"net/http"
)

const (
	// JSONQueryKey is the query key for JSONHandler
	JSONQueryKey = "field"
	// EnvQueryKey is the query key for EnvsHandler
	EnvQueryKey = "env"
)

// ConfigHandler exposes the configuration struct as JSON fields
func (v *Viewer) ConfigHandler(rw http.ResponseWriter, r *http.Request) {
	if v.config == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")

	if configField := r.URL.Query().Get(JSONQueryKey); configField != "" {
		response := v.EnvNotation(configField)
		if response.Value == nil {
			rw.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(rw).Encode(map[string]string{
				"error": "field not found",
			})
			return
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}

	err := json.NewEncoder(rw).Encode(v.config)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

// DetailedConfigHandler exposes the detailed configuration struct as JSON fields
func (v *Viewer) DetailedConfigHandler(rw http.ResponseWriter, r *http.Request) {
	if v.configMap == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")

	if configField := r.URL.Query().Get(JSONQueryKey); configField != "" {
		response := v.EnvNotation(configField)
		if response.Value == nil {
			rw.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(rw).Encode(map[string]string{
				"error": "field not found",
			})
			if err != nil {
				return
			}
			return
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	err := json.NewEncoder(rw).Encode(v.configMap)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// EnvsHandler expose the environment variables of the configuration struct
func (v *Viewer) EnvsHandler(rw http.ResponseWriter, r *http.Request) {
	if v.envs == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")

	if env := r.URL.Query().Get(EnvQueryKey); env != "" {
		response := v.JSONNotation(env)
		if response.Value == nil {
			rw.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(rw).Encode(map[string]string{
				"error": "environment variable not found",
			})
			if err != nil {
				return
			}
			return
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	err := json.NewEncoder(rw).Encode(v.ParseEnvs())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

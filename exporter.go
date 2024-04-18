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

// JSONHandler exposes the configuration struct as JSON fields
func (v *Viewer) JSONHandler(rw http.ResponseWriter, r *http.Request) {
	if v.configMap == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if configField := r.URL.Query().Get(JSONQueryKey); configField != "" {
		err := json.NewEncoder(rw).Encode(v.EnvNotation(configField))
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
	if v.config == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)

	if env := r.URL.Query().Get(EnvQueryKey); env != "" {
		err := json.NewEncoder(rw).Encode(v.JSONNotation(env))
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

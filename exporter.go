package struct_viewer

import (
	"encoding/json"
	"net/http"
)

// JSONHandler exposes the configuration struct as JSON fields
func (v *Viewer) JSONHandler(rw http.ResponseWriter, r *http.Request) {
	if v.config == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)

	err := json.NewEncoder(rw).Encode(v.config)
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

	err := json.NewEncoder(rw).Encode(v.ParseEnvs())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

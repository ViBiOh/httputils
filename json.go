package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type results struct {
	Results interface{} `json:"results"`
}

// ResponseJSON write marshalled obj to http.ResponseWriter with correct header
func ResponseJSON(w http.ResponseWriter, obj interface{}) {
	if objJSON, err := json.Marshal(obj); err == nil {
		w.Header().Add(`Content-Type`, `application/json`)
		w.Header().Add(`Cache-Control`, `no-cache`)
		w.Write(objJSON)
	} else {
		InternalServer(w, fmt.Errorf(`Error while marshalling JSON response: %v`, err))
	}
}

// ResponseArrayJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponseArrayJSON(w http.ResponseWriter, array interface{}) {
	ResponseJSON(w, results{array})
}

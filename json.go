package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type results struct {
	Results interface{} `json:"results"`
}

type pagination struct {
	Results interface{} `json:"results"`
	Total   int64       `json:"total"`
}

// ResponseJSON write marshalled obj to http.ResponseWriter with correct header
func ResponseJSON(w http.ResponseWriter, status int, obj interface{}) {
	if objJSON, err := json.Marshal(obj); err == nil {
		w.Header().Add(`Content-Type`, `application/json`)
		w.Header().Add(`Cache-Control`, `no-cache`)
		w.WriteHeader(status)
		w.Write(objJSON)
	} else {
		InternalServer(w, fmt.Errorf(`Error while marshalling JSON response: %v`, err))
	}
}

// ResponseArrayJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponseArrayJSON(w http.ResponseWriter, status int, array interface{}) {
	ResponseJSON(w, status, results{array})
}

// ResponsPaginatedJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponsPaginatedJSON(w http.ResponseWriter, status int, total int64, array interface{}) {
	ResponseJSON(w, status, pagination{Results: array, Total: total})
}

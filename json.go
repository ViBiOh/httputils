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
	Results   interface{} `json:"results"`
	Page      int64       `json:"page"`
	PageSize  int64       `json:"pageSize"`
	PageCount int64       `json:"pageCount"`
	Total     int64       `json:"total"`
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

// ResponsePaginatedJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponsePaginatedJSON(w http.ResponseWriter, status int, page int64, pageSize int64, total int64, array interface{}) {
	pageCount := int64(total / pageSize)
	if total%pageSize != 0 {
		pageCount++
	}

	ResponseJSON(w, status, pagination{Results: array, Page: page, PageSize: pageSize, PageCount: pageCount, Total: total})
}

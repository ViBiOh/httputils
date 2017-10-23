package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

// IsPretty determine if pretty is defined in query params
func IsPretty(rawQuery string) (pretty bool) {
	if params, err := url.ParseQuery(rawQuery); err == nil {
		if _, ok := params[`pretty`]; ok {
			pretty = true
		}
	}

	return
}

// ResponseJSON write marshalled obj to http.ResponseWriter with correct header
func ResponseJSON(w http.ResponseWriter, status int, obj interface{}, pretty bool) {
	var objJSON []byte
	var err error

	if pretty {
		objJSON, err = json.MarshalIndent(obj, ``, `  `)
	} else {
		objJSON, err = json.Marshal(obj)
	}

	if err == nil {
		w.Header().Add(`Content-Type`, `application/json`)
		w.Header().Add(`Cache-Control`, `no-cache`)
		w.WriteHeader(status)
		w.Write(objJSON)
	} else {
		InternalServer(w, fmt.Errorf(`Error while marshalling JSON response: %v`, err))
	}
}

// ResponseArrayJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponseArrayJSON(w http.ResponseWriter, status int, array interface{}, pretty bool) {
	ResponseJSON(w, status, results{array}, pretty)
}

// ResponsePaginatedJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponsePaginatedJSON(w http.ResponseWriter, status int, page int64, pageSize int64, total int64, array interface{}, pretty bool) {
	pageCount := int64(total / pageSize)
	if total%pageSize != 0 {
		pageCount++
	}

	ResponseJSON(w, status, pagination{Results: array, Page: page, PageSize: pageSize, PageCount: pageCount, Total: total}, pretty)
}

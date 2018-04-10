package httpjson

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type results struct {
	Results interface{} `json:"results"`
}

type pagination struct {
	Results   interface{} `json:"results"`
	Page      uint        `json:"page"`
	PageSize  uint        `json:"pageSize"`
	PageCount uint        `json:"pageCount"`
	Total     uint        `json:"total"`
}

// IsPretty determine if pretty is defined in query params
func IsPretty(rawQuery string) bool {
	if params, err := url.ParseQuery(rawQuery); err == nil {
		if _, ok := params[`pretty`]; ok {
			if pretty, err := strconv.ParseBool(params[`pretty`][0]); err == nil {
				return pretty
			}
			return true
		}
	}

	return false
}

// ResponseJSON write marshalled obj to http.ResponseWriter with correct header
func ResponseJSON(w http.ResponseWriter, status int, obj interface{}, pretty bool) error {
	var objJSON []byte
	var err error

	if pretty {
		objJSON, err = json.MarshalIndent(obj, ``, `  `)
	} else {
		objJSON, err = json.Marshal(obj)
	}

	if err != nil {
		return fmt.Errorf(`Error while marshalling JSON response: %v`, err)
	}

	w.Header().Set(`Content-Type`, `application/json; charset=utf-8`)
	w.Header().Set(`Cache-Control`, `no-cache`)
	w.WriteHeader(status)

	if _, err := w.Write(objJSON); err != nil {
		return fmt.Errorf(`Error while writing JSON: %v`, err)
	}
	return nil
}

// ResponseArrayJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponseArrayJSON(w http.ResponseWriter, status int, array interface{}, pretty bool) error {
	return ResponseJSON(w, status, results{array}, pretty)
}

// ResponsePaginatedJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponsePaginatedJSON(w http.ResponseWriter, status int, page uint, pageSize uint, total uint, array interface{}, pretty bool) error {
	pageCount := uint(total / pageSize)
	if total%pageSize != 0 {
		pageCount++
	}

	return ResponseJSON(w, status, pagination{Results: array, Page: page, PageSize: pageSize, PageCount: pageCount, Total: total}, pretty)
}

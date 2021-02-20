package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/query"
)

var (
	// ErrCannotMarshall occurs when marshaller failed
	ErrCannotMarshall = errors.New("cannot marshall json")
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
func IsPretty(r *http.Request) bool {
	return query.GetBool(r, "pretty")
}

// ResponseJSON write marshalled obj to http.ResponseWriter with correct header
func ResponseJSON(w http.ResponseWriter, status int, obj interface{}, pretty bool) {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)

	if err := encoder.Encode(obj); err != nil {
		httperror.InternalServerError(w, fmt.Errorf("%s: %w", err, ErrCannotMarshall))
	}
}

// ResponseArrayJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponseArrayJSON(w http.ResponseWriter, status int, array interface{}, pretty bool) {
	ResponseJSON(w, status, results{array}, pretty)
}

// ResponsePaginatedJSON write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func ResponsePaginatedJSON(w http.ResponseWriter, status int, page uint, pageSize uint, total uint, array interface{}, pretty bool) {
	pageCount := total / pageSize
	if total%pageSize != 0 {
		pageCount++
	}

	ResponseJSON(w, status, pagination{Results: array, Page: page, PageSize: pageSize, PageCount: pageCount, Total: total}, pretty)
}

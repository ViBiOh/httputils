package query

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	// ErrInvalidValue occurs when value is in invalid format
	ErrInvalidValue = errors.New("invalid value")

	// ErrMaxPageSizeExceeded occurs when pagesize read is above defined limit
	ErrMaxPageSizeExceeded = errors.New("maximum page size exceeded")

	// ErrPageSizeInvalid occurs when pagesize read is equal to 0
	ErrPageSizeInvalid = errors.New("page size must be greater than zero")
)

// Pagination describes pagination params
type Pagination struct {
	LastKey  string
	Sort     string
	PageSize uint
	Desc     bool
}

// ParsePagination parse common pagination param from request
func ParsePagination(r *http.Request, defaultPageSize, maxPageSize uint) (pagination Pagination, err error) {
	var parsed uint64
	var parsedUint uint
	var params url.Values

	params, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		err = fmt.Errorf("%s: %w", err, ErrInvalidValue)
		return
	}

	pagination.PageSize = defaultPageSize
	rawPageSize := strings.TrimSpace(params.Get("pageSize"))
	if len(rawPageSize) != 0 {
		parsed, err = strconv.ParseUint(rawPageSize, 10, 32)
		parsedUint = uint(parsed)
		if err != nil {
			err = fmt.Errorf("pageSize is invalid %s: %w", err, ErrInvalidValue)
			return
		}

		if parsedUint > maxPageSize {
			err = ErrMaxPageSizeExceeded
			return
		}

		if parsedUint < 1 {
			err = ErrPageSizeInvalid
			return
		}

		pagination.PageSize = parsedUint
	}

	pagination.Sort = ""
	rawSortKey := strings.TrimSpace(params.Get("sort"))
	if len(rawSortKey) != 0 {
		pagination.Sort = rawSortKey
	}

	pagination.Desc = GetBool(r, "desc")
	pagination.LastKey = strings.TrimSpace(params.Get("lastKey"))

	return
}

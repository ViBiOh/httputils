package pagination

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
)

var (
	// ErrMaxPageSizeExceeded occurs when pagesize read is above defined limit
	ErrMaxPageSizeExceeded = errors.New("maximum page size exceeded")

	// ErrPageSizeEqualZero occurs when pagesize read is equal to 0
	ErrPageSizeEqualZero = errors.New("page size must be greater than zero")
)

// ParseParams parse common pagination param from request
func ParseParams(r *http.Request, defaultPage, defaultPageSize, maxPageSize uint) (page, pageSize uint, sortKey string, sortAsc bool, err error) {
	var parsed uint64
	var parsedUint uint
	var params url.Values

	params, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return
	}

	page = defaultPage
	rawPage := strings.TrimSpace(params.Get("page"))
	if rawPage != "" {
		parsed, err = strconv.ParseUint(rawPage, 10, 32)
		if err != nil {
			err = errors.New("page is invalid %#v", err)
			return
		}

		page = uint(parsed)
	}

	pageSize = defaultPageSize
	rawPageSize := strings.TrimSpace(params.Get("pageSize"))
	if rawPageSize != "" {
		parsed, err = strconv.ParseUint(rawPageSize, 10, 32)
		parsedUint = uint(parsed)
		if err != nil {
			err = errors.New("pageSize is invalid %#v", err)
			return
		}

		if parsedUint > maxPageSize {
			err = errors.WithStack(ErrMaxPageSizeExceeded)
			return
		}

		if parsedUint < 1 {
			err = errors.WithStack(ErrPageSizeEqualZero)
			return
		}

		pageSize = parsedUint
	}

	sortKey = ""
	rawSortKey := strings.TrimSpace(params.Get("sort"))
	if rawSortKey != "" {
		sortKey = rawSortKey
	}

	sortAsc = true
	if rawValue, ok := params["desc"]; ok {
		if len(rawValue) == 0 {
			sortAsc = false
		} else if value, strconvErr := strconv.ParseBool(rawValue[0]); strconvErr != nil {
			err = errors.WithStack(strconvErr)
			return
		} else {
			sortAsc = !value
		}
	}

	return
}

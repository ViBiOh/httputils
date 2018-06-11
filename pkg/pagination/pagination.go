package pagination

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ErrMaxPageSizeExceeded occurs when pagesize read is above defined limit
var ErrMaxPageSizeExceeded = errors.New(`Maximum page size exceeded`)

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
	rawPage := strings.TrimSpace(params.Get(`page`))
	if rawPage != `` {
		parsed, err = strconv.ParseUint(rawPage, 10, 32)
		parsedUint = uint(parsed)
		if err != nil {
			err = fmt.Errorf(`Error while parsing page param: %v`, err)
			return
		}

		page = parsedUint
	}

	pageSize = defaultPageSize
	rawPageSize := strings.TrimSpace(params.Get(`pageSize`))
	if rawPageSize != `` {
		parsed, err = strconv.ParseUint(rawPageSize, 10, 32)
		parsedUint = uint(parsed)
		if err != nil {
			err = fmt.Errorf(`Error while parsing pageSize param: %v`, err)
			return
		} else if parsedUint > maxPageSize {
			err = ErrMaxPageSizeExceeded
			return
		}

		pageSize = parsedUint
	}

	sortKey = ``
	rawSortKey := strings.TrimSpace(params.Get(`sort`))
	if rawSortKey != `` {
		sortKey = rawSortKey
	}

	sortAsc = true
	if _, ok := params[`desc`]; ok {
		sortAsc = false
	}

	return
}

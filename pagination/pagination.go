package pagination

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ParsePaginationParams parse common pagination param from request
func ParsePaginationParams(r *http.Request, defaultPageSize, maxPageSize int64) (page, pageSize int64, sortKey string, sortAsc bool, err error) {
	var parsedInt int64
	var params url.Values

	params, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return
	}

	page = 1
	rawPage := params.Get(`page`)
	if rawPage != `` {
		parsedInt, err = strconv.ParseInt(rawPage, 10, 64)
		if err != nil {
			err = fmt.Errorf(`Error while parsing page param: %v`, err)
			return
		}

		page = parsedInt
	}

	pageSize = defaultPageSize
	rawPageSize := params.Get(`pageSize`)
	if rawPageSize != `` {
		parsedInt, err = strconv.ParseInt(rawPageSize, 10, 64)
		if err != nil {
			err = fmt.Errorf(`Error while parsing pageSize param: %v`, err)
			return
		} else if parsedInt > maxPageSize {
			err = fmt.Errorf(`Maximum page size exceeded`)
			return
		}

		pageSize = parsedInt
	}

	sortKey = ``
	rawSortKey := params.Get(`sort`)
	if rawSortKey != `` {
		sortKey = rawSortKey
	}

	sortAsc = true
	if _, ok := params[`desc`]; ok {
		sortAsc = false
	}

	return
}

package query

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrInvalidValue        = errors.New("invalid value")
	ErrMaxPageSizeExceeded = errors.New("maximum page size exceeded")
	ErrPageSizeInvalid     = errors.New("page size must be greater than zero")
)

type Pagination struct {
	Last     string
	Sort     string
	PageSize uint
	Desc     bool
}

func (p Pagination) LinkNextHeader(urlPath string, extraArgs url.Values) string {
	data := url.Values{}
	for key, values := range extraArgs {
		for _, value := range values {
			data.Add(key, value)
		}
	}

	data.Add("pageSize", fmt.Sprintf("%d", p.PageSize))
	data.Add("last", p.Last)

	if len(p.Sort) != 0 {
		data.Add("sort", p.Sort)
	}

	if p.Desc {
		data.Add("desc", strconv.FormatBool(p.Desc))
	}

	return fmt.Sprintf(`<%s?%s>; rel="next"`, urlPath, data.Encode())
}

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
	rawPageSize := params.Get("pageSize")
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
	rawSortKey := params.Get("sort")
	if len(rawSortKey) != 0 {
		pagination.Sort = rawSortKey
	}

	pagination.Desc = GetBool(r, "desc")
	pagination.Last = params.Get("last")

	return
}

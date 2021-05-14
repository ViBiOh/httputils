package query

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestParsePagination(t *testing.T) {
	var cases = []struct {
		intention       string
		request         *http.Request
		defaultPageSize uint
		maxPageSize     uint
		want            Pagination
		wantErr         error
	}{
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
			},
			nil,
		},
		{
			"simple value",
			httptest.NewRequest(http.MethodGet, "/?pageSize=50&sort=name&desc", nil),
			20,
			100,
			Pagination{
				PageSize: 50,
				Sort:     "name",
				Desc:     true,
			},
			nil,
		},
		{
			"invalid pageSize",
			httptest.NewRequest(http.MethodGet, "/?pageSize=invalid", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
			},
			ErrInvalidValue,
		},
		{
			"too high pageSize",
			httptest.NewRequest(http.MethodGet, "/?pageSize=150", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
			},
			ErrMaxPageSizeExceeded,
		},
		{
			"invalid pageSize number",
			httptest.NewRequest(http.MethodGet, "/?pageSize=0", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
			},
			ErrPageSizeInvalid,
		},
		{
			"sort",
			httptest.NewRequest(http.MethodGet, "/?sort=name", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
				Sort:     "name",
			},
			nil,
		},
		{
			"order",
			httptest.NewRequest(http.MethodGet, "/?desc", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
				Desc:     true,
			},
			nil,
		},
		{
			"order with value",
			httptest.NewRequest(http.MethodGet, "/?desc=false", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
			},
			nil,
		},
		{
			"last key",
			httptest.NewRequest(http.MethodGet, "/?lastKey=8000", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
				LastKey:  "8000",
			},
			nil,
		},
		{
			"error",
			&http.Request{
				URL: &url.URL{
					RawQuery: "/%1",
				},
			},
			0,
			0,
			Pagination{},
			ErrInvalidValue,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			result, err := ParsePagination(tc.request, tc.defaultPageSize, tc.maxPageSize)

			failed := false

			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("ParsePagination() = (%#v, `%s`), want (%#v, `%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

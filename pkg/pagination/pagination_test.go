package pagination

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseParams(t *testing.T) {
	var cases = []struct {
		intention       string
		request         *http.Request
		defaultPage     uint
		defaultPageSize uint
		maxPageSize     uint
		want            Pagination
		wantErr         error
	}{
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
			},
			nil,
		},
		{
			"invalid page",
			httptest.NewRequest(http.MethodGet, "/?page=invalid", nil),
			1,
			20,
			100,
			Pagination{
				Page: 1,
			},
			ErrInvalidValue,
		},
		{
			"invalid pageSize",
			httptest.NewRequest(http.MethodGet, "/?pageSize=invalid", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
			},
			ErrInvalidValue,
		},
		{
			"too high pageSize",
			httptest.NewRequest(http.MethodGet, "/?pageSize=150", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
			},
			ErrMaxPageSizeExceeded,
		},
		{
			"invalid pageSize number",
			httptest.NewRequest(http.MethodGet, "/?pageSize=0", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
			},
			ErrPageSizeInvalid,
		},
		{
			"sort",
			httptest.NewRequest(http.MethodGet, "/?sort=name", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
				Sort:     "name",
			},
			nil,
		},
		{
			"order",
			httptest.NewRequest(http.MethodGet, "/?desc", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
				Desc:     true,
			},
			nil,
		},
		{
			"order with value",
			httptest.NewRequest(http.MethodGet, "/?desc=false", nil),
			1,
			20,
			100,
			Pagination{
				Page:     1,
				PageSize: 20,
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := ParseParams(testCase.request, testCase.defaultPage, testCase.defaultPageSize, testCase.maxPageSize)

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ParseParams() = (%#v, %#v), want (%#v, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

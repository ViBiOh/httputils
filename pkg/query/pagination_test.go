package query

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestLinkNextHeader(t *testing.T) {
	type args struct {
		urlPath   string
		extraArgs url.Values
	}

	var cases = []struct {
		intention string
		instance  Pagination
		args      args
		want      string
	}{
		{
			"empty",
			Pagination{
				Last:     "8000",
				PageSize: 20,
			},
			args{
				urlPath: "/list",
			},
			`</list?last=8000&pageSize=20>; rel="next"`,
		},
		{
			"extra empty",
			Pagination{
				Last:     "8000",
				PageSize: 20,
				Sort:     "id",
				Desc:     true,
			},
			args{
				urlPath:   "/list",
				extraArgs: url.Values{},
			},
			`</list?desc=true&last=8000&pageSize=20&sort=id>; rel="next"`,
		},
		{
			"extra args",
			Pagination{
				Last:     "8000",
				PageSize: 20,
			},
			args{
				urlPath: "/list",
				extraArgs: url.Values{
					"query": []string{"search"},
				},
			},
			`</list?last=8000&pageSize=20&query=search>; rel="next"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.LinkNextHeader(tc.args.urlPath, tc.args.extraArgs); got != tc.want {
				t.Errorf("LinkNextHeader() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

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
			httptest.NewRequest(http.MethodGet, "/?last=8000", nil),
			20,
			100,
			Pagination{
				PageSize: 20,
				Last:     "8000",
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

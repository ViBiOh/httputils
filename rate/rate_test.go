package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ViBiOh/httputils"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		want      map[string]interface{}
	}{
		{
			`default prefix`,
			``,
			map[string]interface{}{
				`limit`: nil,
			},
		},
		{
			`given prefix`,
			`test`,
			map[string]interface{}{
				`limit`: nil,
			},
		},
	}

	for _, testCase := range cases {
		if result := Flags(testCase.prefix); len(result) != len(testCase.want) {
			t.Errorf("%v\nFlags(%v) = %v, want %v", testCase.intention, testCase.prefix, result, testCase.want)
		}
	}
}

func Test_checkRate(t *testing.T) {
	var cases = []struct {
		ipRate map[string]uint
		want   bool
	}{
		{
			map[string]uint{},
			true,
		},
		{
			map[string]uint{
				`localhost`: 100,
			},
			true,
		},
		{
			map[string]uint{
				`localhost`: defaultLimit,
			},
			false,
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `/test`, nil)
		request.RemoteAddr = `localhost`

		ipRate = testCase.ipRate

		if result := checkRate(request, defaultLimit); result != testCase.want {
			t.Errorf(`checkRate(%v) = (%v), want (%v)`, testCase.ipRate, result, testCase.want)
		}
	}
}

func Benchmark_checkRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	for i := 0; i < b.N; i++ {
		checkRate(request, defaultLimit)
	}
}

func Test_ServeHTTP(t *testing.T) {
	limit := uint(20)

	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.Header.Set(httputils.ForwardedForHeader, `localhost`)

	calls := make([]time.Time, defaultLimit)
	for i := uint(0); i < defaultLimit; i++ {
		calls[i] = time.Now()
	}

	var cases = []struct {
		request *http.Request
		config  map[string]interface{}
		ipRate  map[string]uint
		want    int
	}{
		{
			request,
			nil,
			map[string]uint{},
			http.StatusOK,
		},
		{
			request,
			map[string]interface{}{
				`limit`: &limit,
			},
			map[string]uint{
				`localhost`: limit,
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			nil,
			map[string]uint{
				`localhost`: defaultLimit - 1,
			},
			http.StatusOK,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits/`, nil),
			nil,
			map[string]uint{
				`localhost`: defaultLimit - 1,
			},
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		ipRate = testCase.ipRate

		response := httptest.NewRecorder()
		Handler(testCase.config, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(response, testCase.request)

		if result := response.Result().StatusCode; result != testCase.want {
			t.Errorf(`ServeHTTP() = (%v) want %v, with ipRate = %v`, result, testCase.want, testCase.ipRate)
		}
	}
}

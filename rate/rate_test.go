package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestGetIP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/`, nil)
	request.RemoteAddr = `localhost`

	requestWithProxy := httptest.NewRequest(http.MethodGet, `/`, nil)
	requestWithProxy.RemoteAddr = `localhost`
	requestWithProxy.Header.Add(forwardedForHeader, `proxy`)

	var cases = []struct {
		r    *http.Request
		want string
	}{
		{
			request,
			`localhost`,
		},
		{
			requestWithProxy,
			`proxy`,
		},
	}

	for _, testCase := range cases {
		if result := getIP(testCase.r); result != testCase.want {
			t.Errorf(`getIP(%v) = %v, want %v`, testCase.r, result, testCase.want)
		}
	}
}

func TestCheckRate(t *testing.T) {
	var cases = []struct {
		ipRate map[string]int
		want   bool
	}{
		{
			map[string]int{},
			true,
		},
		{
			map[string]int{
				`localhost`: 100,
			},
			true,
		},
		{
			map[string]int{
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

func BenchmarkCheckRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	for i := 0; i < b.N; i++ {
		checkRate(request, defaultLimit)
	}
}

func TestServeHTTP(t *testing.T) {
	limit := 20

	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.Header.Add(forwardedForHeader, `localhost`)

	calls := make([]time.Time, defaultLimit)
	for i := 0; i < defaultLimit; i++ {
		calls[i] = time.Now()
	}

	var cases = []struct {
		request *http.Request
		config  map[string]interface{}
		ipRate  map[string]int
		want    int
	}{
		{
			request,
			nil,
			map[string]int{},
			http.StatusOK,
		},
		{
			request,
			map[string]interface{}{
				`limit`: &limit,
			},
			map[string]int{
				`localhost`: limit,
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			nil,
			map[string]int{
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

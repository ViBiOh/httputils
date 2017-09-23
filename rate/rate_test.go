package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
				`localhost`: *ipRateLimit,
			},
			false,
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `/test`, nil)
		request.RemoteAddr = `localhost`

		ipRate = testCase.ipRate

		if result := checkRate(request); result != testCase.want {
			t.Errorf(`checkRate(%v) = (%v), want (%v)`, testCase.ipRate, result, testCase.want)
		}
	}
}

func BenchmarkCheckRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	for i := 0; i < b.N; i++ {
		if result := checkRate(request); result != (ipRate[`localhost`] < *ipRateLimit) {
			b.Errorf(`checkRate() = (%v), want (%v)`, result, ipRate[`localhost`] < *ipRateLimit)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.Header.Add(forwardedForHeader, `localhost`)

	calls := make([]time.Time, *ipRateLimit)
	for i := 0; i < *ipRateLimit; i++ {
		calls[i] = time.Now()
	}

	var cases = []struct {
		request *http.Request
		ipRate  map[string]int
		want    int
	}{
		{
			request,
			map[string]int{},
			http.StatusOK,
		},
		{
			request,
			map[string]int{
				`localhost`: *ipRateLimit,
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			map[string]int{
				`localhost`: *ipRateLimit - 1,
			},
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		ipRate = testCase.ipRate

		response := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(response, testCase.request)

		if result := response.Result().StatusCode; result != testCase.want {
			t.Errorf(`ServeHTTP() = (%v) want %v, with ipRate = %v`, result, testCase.want, testCase.ipRate)
		}
	}
}

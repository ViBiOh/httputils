package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckRate(t *testing.T) {
	calls := make([]time.Time, *ipRateLimit)
	for i := 0; i < *ipRateLimit; i++ {
		calls[i] = time.Now()
	}

	var cases = []struct {
		userRate        map[string]*rateLimit
		forwardedHeader string
		want            bool
	}{
		{
			map[string]*rateLimit{},
			``,
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					calls: []time.Time{
						time.Now(),
					},
				},
			},
			``,
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					calls: []time.Time{
						time.Now().Add(-180 * time.Second),
						time.Now().Add(-90 * time.Second),
						time.Now().Add(-60 * time.Second),
						time.Now().Add(-30 * time.Second),
					},
				},
			},
			``,
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					calls: calls,
				},
			},
			``,
			false,
		},
		{
			map[string]*rateLimit{
				`real-ip`: {
					calls: calls,
				},
			},
			`real-ip`,
			false,
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `/test`, nil)
		request.RemoteAddr = `localhost`
		request.Header.Add(forwardedForHeader, testCase.forwardedHeader)

		userRate = testCase.userRate

		if result := checkRate(request); result != testCase.want {
			t.Errorf(`checkRate(%v) = (%v), want (%v)`, testCase.userRate, result, testCase.want)
		}
	}
}

func BenchmarkCheckRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateLimit)
	for i := 0; i < *ipRateLimit; i++ {
		calls[i] = time.Now()
	}

	var testCase = struct {
		userRate map[string]*rateLimit
		want     bool
	}{
		map[string]*rateLimit{
			`localhost`: {
				calls: calls,
			},
		},
		false,
	}

	for i := 0; i < b.N; i++ {
		userRate = testCase.userRate

		if result := checkRate(request); result != testCase.want {
			b.Errorf(`checkRate(%v) = (%v), want (%v)`, testCase.userRate, result, testCase.want)
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
		request  *http.Request
		userRate map[string]*rateLimit
		want     int
	}{
		{
			request,
			map[string]*rateLimit{},
			http.StatusOK,
		},
		{
			request,
			map[string]*rateLimit{
				`localhost`: {
					calls: calls,
				},
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			map[string]*rateLimit{
				`localhost`: {
					calls: calls,
				},
			},
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		userRate = testCase.userRate

		response := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(response, testCase.request)

		if result := response.Result().StatusCode; result != testCase.want {
			t.Errorf(`ServeHTTP() = (%v) want %v, with userRate = %v`, result, testCase.want, testCase.userRate)
		}
	}
}

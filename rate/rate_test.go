package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckRate(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateCount)
	for i := 0; i < *ipRateCount; i++ {
		calls[i] = time.Now()
	}

	var tests = []struct {
		userRate map[string]*rateLimit
		want     bool
	}{
		{
			map[string]*rateLimit{},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip: `localhost`,
					calls: []time.Time{
						time.Now(),
					},
				},
			},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip: `localhost`,
					calls: []time.Time{
						time.Now().Add(-180 * time.Second),
						time.Now().Add(-90 * time.Second),
						time.Now().Add(-60 * time.Second),
						time.Now().Add(-30 * time.Second),
					},
				},
			},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip:    `localhost`,
					calls: calls,
				},
			},
			false,
		},
	}

	for _, test := range tests {
		userRate = test.userRate

		if result := checkRate(request); result != test.want {
			t.Errorf(`checkRate(%v) = (%v), want (%v)`, test.userRate, result, test.want)
		}
	}
}

func BenchmarkCheckRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateCount)
	for i := 0; i < *ipRateCount; i++ {
		calls[i] = time.Now()
	}

	var test = struct {
		userRate map[string]*rateLimit
		want     bool
	}{
		map[string]*rateLimit{
			`localhost`: {
				ip:    `localhost`,
				calls: calls,
			},
		},
		false,
	}

	for i := 0; i < b.N; i++ {
		userRate = test.userRate

		if result := checkRate(request); result != test.want {
			b.Errorf(`checkRate(%v) = (%v), want (%v)`, test.userRate, result, test.want)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateCount)
	for i := 0; i < *ipRateCount; i++ {
		calls[i] = time.Now()
	}

	var tests = []struct {
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
					ip:    `localhost`,
					calls: calls,
				},
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			map[string]*rateLimit{
				`localhost`: {
					ip:    `localhost`,
					calls: calls,
				},
			},
			http.StatusOK,
		},
	}

	for _, test := range tests {
		userRate = test.userRate

		response := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(response, test.request)

		if result := response.Result().StatusCode; result != test.want {
			t.Errorf(`ServeHTTP() = (%v) want %v, with userRate = %v`, result, test.want, test.userRate)
		}
	}
}

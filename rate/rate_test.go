package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func generateCalls() []*rateLimit {
	current := time.Now().Add(*ipRateDelay * -2).Unix()
	countPerSecond := *ipRateLimit / int(*ipRateDelay/time.Second) * 2

	rateLimits := make([]*rateLimit, 0)
	for i := 0; i < int(*ipRateDelay/time.Second*2); i++ {
		rateLimits = append(rateLimits, &rateLimit{
			unix:  current,
			Count: countPerSecond,
		})

		current++
	}

	return rateLimits
}

func TestCheckRate(t *testing.T) {
	var cases = []struct {
		userRate        map[string][]*rateLimit
		forwardedHeader string
		want            bool
	}{
		{
			map[string][]*rateLimit{},
			``,
			true,
		},
		{
			map[string][]*rateLimit{
				`localhost`: {
					{
						unix:  time.Now().Unix(),
						Count: 1,
					},
				},
			},
			``,
			true,
		},
		{
			map[string][]*rateLimit{
				`localhost`: {
					{
						unix:  time.Now().Unix(),
						Count: *ipRateLimit,
					},
				},
			},
			``,
			false,
		},
		{
			map[string][]*rateLimit{
				`localhost`: generateCalls(),
			},
			``,
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

	var testCase = struct {
		userRate map[string][]*rateLimit
		want     bool
	}{
		map[string][]*rateLimit{
			`localhost`: generateCalls(),
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
		userRate map[string][]*rateLimit
		want     int
	}{
		{
			request,
			map[string][]*rateLimit{},
			http.StatusOK,
		},
		{
			request,
			map[string][]*rateLimit{
				`localhost`: generateCalls(),
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			map[string][]*rateLimit{
				`localhost`: {
					{
						unix:  time.Now().Unix(),
						Count: *ipRateLimit + 10,
					},
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

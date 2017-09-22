package rate

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
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
			count: countPerSecond,
		})

		current++
	}

	return rateLimits
}

func loadUserRate(toLoad map[string][]*rateLimit) {
	userRate = sync.Map{}
	for key, value := range toLoad {
		userRate.Store(key, value)
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

func TestGetRateLimits(t *testing.T) {
	var cases = []struct {
		userRate       map[string][]*rateLimit
		want           string
		wantRateLimits []*rateLimit
	}{
		{
			nil,
			`localhost`,
			[]*rateLimit{},
		},
		{
			map[string][]*rateLimit{`localhost`: {{1000, 0}}},
			`localhost`,
			[]*rateLimit{{1000, 0}},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `/`, nil)
		request.RemoteAddr = `localhost`

		loadUserRate(testCase.userRate)

		result, rateLimits := getRateLimits(request)

		if result != testCase.want {
			t.Errorf(`getRateLimits() = %v, want %v`, result, testCase.want)
		}

		if !reflect.DeepEqual(rateLimits, testCase.wantRateLimits) {
			t.Errorf(`getRateLimits() = %v, want %v`, rateLimits, testCase.wantRateLimits)
		}
	}
}

func TestCleanRateLimits(t *testing.T) {
	var cases = []struct {
		rateLimits          []*rateLimit
		nowMinusDelaySecond int64
		want                []*rateLimit
	}{
		{
			nil,
			0,
			nil,
		},
		{
			[]*rateLimit{{1000, 0}},
			800,
			[]*rateLimit{{1000, 0}},
		},
		{
			[]*rateLimit{{1000, 0}},
			1200,
			[]*rateLimit{},
		},
	}

	for _, testCase := range cases {
		if result := cleanRateLimits(testCase.rateLimits, testCase.nowMinusDelaySecond); !reflect.DeepEqual(result, testCase.want) {
			t.Errorf(`cleanRateLimits(%v, %v) = %v, want %v`, testCase.rateLimits, testCase.nowMinusDelaySecond, result, testCase.want)
		}
	}
}

func TestCleanUserRate(t *testing.T) {
	now, nowMinusDelay := getUnix()

	var cases = []struct {
		userRate map[string][]*rateLimit
		want     map[string][]*rateLimit
	}{
		{
			nil,
			map[string][]*rateLimit{},
		},
		{
			map[string][]*rateLimit{`localhost`: {{now, 0}}},
			map[string][]*rateLimit{`localhost`: {{now, 0}}},
		},
		{
			map[string][]*rateLimit{`localhost`: {{now, 0}}, `proxy`: {{nowMinusDelay - 10, 0}}},
			map[string][]*rateLimit{`localhost`: {{now, 0}}},
		},
	}

	for _, testCase := range cases {
		loadUserRate(testCase.userRate)
		cleanUserRate()

		result := make(map[string][]*rateLimit, 0)
		userRate.Range(func(key, value interface{}) bool {
			result[key.(string)] = value.([]*rateLimit)
			return true
		})

		if !reflect.DeepEqual(result, testCase.want) {
			t.Errorf(`cleanUserRate() = %v, want %v`, result, testCase.want)
		}
	}
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
						count: 1,
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
						count: *ipRateLimit,
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

		loadUserRate(testCase.userRate)

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

	loadUserRate(testCase.userRate)

	for i := 0; i < b.N; i++ {
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
						count: *ipRateLimit + 10,
					},
				},
			},
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		loadUserRate(testCase.userRate)

		response := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(response, testCase.request)

		if result := response.Result().StatusCode; result != testCase.want {
			t.Errorf(`ServeHTTP() = (%v) want %v, with userRate = %v`, result, testCase.want, testCase.userRate)
		}
	}
}

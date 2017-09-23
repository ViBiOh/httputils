package rate

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func generateCalls() []*rate {
	current := time.Now().Add(*ipRateDelay * -2).Unix()
	countPerSecond := *ipRateLimit / int(*ipRateDelay/time.Second) * 2

	rates := make([]*rate, 0)
	for i := 0; i < int(*ipRateDelay/time.Second*2); i++ {
		rates = append(rates, &rate{
			unix:  current,
			count: countPerSecond,
		})

		current++
	}

	return rates
}

func loadUserRate(toLoad map[string][]*rate) {
	ipRate = sync.Map{}
	for key, value := range toLoad {
		ipRate.Store(key, value)
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

func TestGetRates(t *testing.T) {
	var cases = []struct {
		userRate  map[string][]*rate
		want      string
		wantRates []*rate
	}{
		{
			nil,
			`localhost`,
			[]*rate{},
		},
		{
			map[string][]*rate{`localhost`: {{1000, 0}}},
			`localhost`,
			[]*rate{{1000, 0}},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `/`, nil)
		request.RemoteAddr = `localhost`

		loadUserRate(testCase.userRate)

		result, rates := getRates(request)

		if result != testCase.want {
			t.Errorf(`getRates() = %v, want %v`, result, testCase.want)
		}

		if !reflect.DeepEqual(rates, testCase.wantRates) {
			t.Errorf(`getRates() = %v, want %v`, rates, testCase.wantRates)
		}
	}
}

func TestCleanRates(t *testing.T) {
	var cases = []struct {
		rates               []*rate
		nowMinusDelaySecond int64
		want                []*rate
	}{
		{
			nil,
			0,
			nil,
		},
		{
			[]*rate{{1000, 0}},
			800,
			[]*rate{{1000, 0}},
		},
		{
			[]*rate{{1000, 0}},
			1200,
			[]*rate{},
		},
	}

	for _, testCase := range cases {
		if result := cleanRates(testCase.rates, testCase.nowMinusDelaySecond); !reflect.DeepEqual(result, testCase.want) {
			t.Errorf(`cleanRates(%v, %v) = %v, want %v`, testCase.rates, testCase.nowMinusDelaySecond, result, testCase.want)
		}
	}
}

func TestCleanIPRate(t *testing.T) {
	now, nowMinusDelay := getUnix()

	var cases = []struct {
		userRate map[string][]*rate
		want     map[string][]*rate
	}{
		{
			nil,
			map[string][]*rate{},
		},
		{
			map[string][]*rate{`localhost`: {{now, 0}}},
			map[string][]*rate{`localhost`: {{now, 0}}},
		},
		{
			map[string][]*rate{`localhost`: {{now, 0}}, `proxy`: {{nowMinusDelay - 10, 0}}},
			map[string][]*rate{`localhost`: {{now, 0}}},
		},
	}

	for _, testCase := range cases {
		loadUserRate(testCase.userRate)
		cleanIPRate()

		result := make(map[string][]*rate, 0)
		ipRate.Range(func(key, value interface{}) bool {
			result[key.(string)] = value.([]*rate)
			return true
		})

		if !reflect.DeepEqual(result, testCase.want) {
			t.Errorf(`cleanIPRate() = %v, want %v`, result, testCase.want)
		}
	}
}

func TestCheckRate(t *testing.T) {
	var cases = []struct {
		userRate        map[string][]*rate
		forwardedHeader string
		want            bool
	}{
		{
			map[string][]*rate{},
			``,
			true,
		},
		{
			map[string][]*rate{
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
			map[string][]*rate{
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
			map[string][]*rate{
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
		userRate map[string][]*rate
		want     bool
	}{
		map[string][]*rate{
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
		userRate map[string][]*rate
		want     int
	}{
		{
			request,
			map[string][]*rate{},
			http.StatusOK,
		},
		{
			request,
			map[string][]*rate{
				`localhost`: generateCalls(),
			},
			http.StatusTooManyRequests,
		},
		{
			httptest.NewRequest(http.MethodGet, `/rate_limits`, nil),
			map[string][]*rate{
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

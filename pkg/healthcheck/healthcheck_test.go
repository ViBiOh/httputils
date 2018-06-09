package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

func Test_Handler(t *testing.T) {
	subHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	var cases = []struct {
		intention  string
		request    *http.Request
		app        *App
		want       string
		wantStatus int
	}{
		{
			`should reject non-GET method`,
			httptest.NewRequest(http.MethodOptions, `/`, nil),
			NewApp(),
			``,
			http.StatusMethodNotAllowed,
		},
		{
			`should handle closed Handler`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			NewApp().Close(),
			``,
			http.StatusServiceUnavailable,
		},
		{
			`should handle nil subHandler`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			NewApp(),
			``,
			http.StatusOK,
		},
		{
			`should handle call given subHandler`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			NewApp().NextHealthcheck(subHandler),
			``,
			http.StatusTeapot,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()

		testCase.app.Handler().ServeHTTP(writer, testCase.request)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nHandler(%+v) = %+v, want status %+v", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nHandler(%+v) = %+v, want %+v", testCase.intention, testCase.request, string(result), testCase.want)
		}
	}
}

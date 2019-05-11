package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

func TestHandler(t *testing.T) {
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
			"should reject non-GET method",
			httptest.NewRequest(http.MethodOptions, "/", nil),
			New(),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"should handle closed Handler",
			httptest.NewRequest(http.MethodGet, "/", nil),
			New().Close(),
			"",
			http.StatusServiceUnavailable,
		},
		{
			"should handle nil subHandler",
			httptest.NewRequest(http.MethodGet, "/", nil),
			New(),
			"",
			http.StatusNoContent,
		},
		{
			"should handle call given subHandler",
			httptest.NewRequest(http.MethodGet, "/", nil),
			New().NextHealthcheck(subHandler),
			"",
			http.StatusTeapot,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			testCase.app.Handler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Handler(%+v) = %+v, want status %+v", testCase.request, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Handler(%+v) = %+v, want %+v", testCase.request, string(result), testCase.want)
			}
		})
	}
}

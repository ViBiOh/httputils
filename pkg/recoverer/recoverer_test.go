package recoverer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	failingHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nilMap map[string]string

		nilMap["fail"] = "yes"

		w.WriteHeader(http.StatusOK)
	})
)

func TestMiddleware(t *testing.T) {
	var cases = []struct {
		intention  string
		next       http.Handler
		request    *http.Request
		wantStatus int
	}{
		{
			"success",
			handler,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
		},
		{
			"fail",
			failingHandler,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Middleware(tc.next).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, tc.wantStatus)
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	middleware := Middleware(nil)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

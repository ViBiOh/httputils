package httprecover

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

		nilMap["fail"] = "yes" //nolint:staticcheck

		w.WriteHeader(http.StatusOK)
	})
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		next       http.Handler
		request    *http.Request
		wantStatus int
	}{
		"success": {
			handler,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
		},
		"fail": {
			failingHandler,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusInternalServerError,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Middleware(testCase.next).ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, testCase.wantStatus)
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	middleware := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for b.Loop() {
		middleware.ServeHTTP(writer, request)
	}
}

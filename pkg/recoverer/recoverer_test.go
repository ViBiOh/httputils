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

		nilMap["fail"] = "yes" //nolint:staticcheck

		w.WriteHeader(http.StatusOK)
	})
)

func TestMiddleware(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Middleware(tc.next).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, tc.wantStatus)
			}
		})
	}
}

func TestLoggerRecoverer(t *testing.T) {
	cases := map[string]struct{}{
		"simple": {},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			func() {
				defer LoggerRecover()

				panic("catch me if you can")
			}()
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	middleware := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

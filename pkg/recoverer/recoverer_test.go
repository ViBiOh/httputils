package recoverer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		intention, testCase := intention, testCase

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

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		func() {
			defer Error(nil)

			panic("catch me if you can")
		}()
	})

	t.Run("valued", func(t *testing.T) {
		t.Parallel()

		var err error

		func() {
			defer Error(&err)

			panic("catch me if you can")
		}()

		assert.ErrorContains(t, err, "recovered")
	})

	t.Run("join", func(t *testing.T) {
		t.Parallel()

		err := errors.New("invalid")

		func() {
			defer Error(&err)

			panic("catch me if you can")
		}()

		assert.ErrorContains(t, err, "invalid")
		assert.ErrorContains(t, err, "recovered")
	})
}

func TestHandler(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{}{
		"simple": {},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			var err error

			func() {
				defer Handler(func(e error) {
					err = e
				})

				panic("catch me if you can")
			}()

			assert.ErrorContains(t, err, "recovered")
		})
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{}{
		"simple": {},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			func() {
				defer Logger()

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

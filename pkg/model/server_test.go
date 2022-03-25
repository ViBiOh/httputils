package model

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func readContent(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	content, err := io.ReadAll(body)

	if closeErr := body.Close(); closeErr != nil {
		err = WrapError(err, closeErr)
	}

	return content, err
}

func TestChainMiddlewares(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("handler")); err != nil {
			t.Errorf("unable to write: %s", err)
		}
	})

	middlewareNotFound := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			next.ServeHTTP(w, r)
		})
	}

	cases := map[string]struct {
		request     *http.Request
		middlewares []Middleware
		want        string
		wantStatus  int
		wantHeader  http.Header
	}{
		"nil chain": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			nil,
			"handler",
			http.StatusOK,
			http.Header{},
		},
		"values": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			[]Middleware{middlewareNotFound},
			"handler",
			http.StatusNotFound,
			http.Header{},
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ChainMiddlewares(handler, tc.middlewares...).ServeHTTP(writer, tc.request)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("ChainMiddlewares = %d, want %d", result, tc.wantStatus)
			}

			if result, _ := readContent(writer.Result().Body); string(result) != tc.want {
				t.Errorf("ChainMiddlewares = `%s`, want `%s`", string(result), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if result := writer.Header().Get(key); result != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, result, want)
				}
			}
		})
	}
}

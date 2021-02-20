package model

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestChainMiddlewares(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})

	middlewareNotFound := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			next.ServeHTTP(w, r)
		})
	}

	var cases = []struct {
		intention   string
		request     *http.Request
		middlewares []Middleware
		want        string
		wantStatus  int
		wantHeader  http.Header
	}{
		{
			"nil chain",
			httptest.NewRequest(http.MethodGet, "/", nil),
			nil,
			"handler",
			http.StatusOK,
			http.Header{},
		},
		{
			"values",
			httptest.NewRequest(http.MethodGet, "/", nil),
			[]Middleware{middlewareNotFound},
			"handler",
			http.StatusNotFound,
			http.Header{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ChainMiddlewares(handler, testCase.middlewares...).ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("ChainMiddlewares = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("ChainMiddlewares = `%s`, want `%s`", string(result), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if result := writer.Header().Get(key); result != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, result, want)
				}
			}
		})
	}
}

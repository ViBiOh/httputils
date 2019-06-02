package cors

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandler(t *testing.T) {
	var cases = []struct {
		intention  string
		app        App
		next       http.Handler
		request    *http.Request
		want       int
		wantHeader http.Header
	}{
		{
			"default param",
			App{
				origin:      "*",
				headers:     "Content-Type",
				methods:     http.MethodGet,
				exposes:     "",
				credentials: "true",
			},
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Access-Control-Allow-Origin":      []string{"*"},
				"Access-Control-Allow-Headers":     []string{"Content-Type"},
				"Access-Control-Allow-Methods":     []string{http.MethodGet},
				"Access-Control-Expose-Headers":    []string{""},
				"Access-Control-Allow-Credentials": []string{"true"},
			},
		},
		{
			"default param",
			App{
				origin:      "*",
				headers:     "Content-Type,Authorization",
				methods:     http.MethodPost,
				exposes:     "",
				credentials: "false",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusNoContent,
			http.Header{
				"Access-Control-Allow-Origin":      []string{"*"},
				"Access-Control-Allow-Headers":     []string{"Content-Type,Authorization"},
				"Access-Control-Allow-Methods":     []string{http.MethodPost},
				"Access-Control-Expose-Headers":    []string{""},
				"Access-Control-Allow-Credentials": []string{"false"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			testCase.app.Handler(testCase.next).ServeHTTP(writer, testCase.request)

			if writer.Code != testCase.want {
				t.Errorf("Handler(%#v) = %d, want %d", testCase.next, writer.Code, testCase.want)
			}

			if !reflect.DeepEqual(writer.Header(), testCase.wantHeader) {
				t.Errorf("Handler(%#v) = %#v, want %#v", testCase.next, writer.Header(), testCase.wantHeader)
			}
		})
	}
}

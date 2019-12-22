package cors

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -credentials\n    \t[cors] Access-Control-Allow-Credentials {SIMPLE_CREDENTIALS}\n  -expose string\n    \t[cors] Access-Control-Expose-Headers {SIMPLE_EXPOSE}\n  -headers string\n    \t[cors] Access-Control-Allow-Headers {SIMPLE_HEADERS} (default \"Content-Type\")\n  -methods string\n    \t[cors] Access-Control-Allow-Methods {SIMPLE_METHODS} (default \"GET\")\n  -origin string\n    \t[cors] Access-Control-Allow-Origin {SIMPLE_ORIGIN} (default \"*\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	var cases = []struct {
		intention string
		want      App
	}{
		{
			"simple",
			&app{
				origin:      "*",
				headers:     "Content-Type",
				methods:     http.MethodGet,
				credentials: "false",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)

			if result := New(Flags(fs, "")); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("New() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
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
			app{
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
			app{
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

			testCase.app.Middleware(testCase.next).ServeHTTP(writer, testCase.request)

			if writer.Code != testCase.want {
				t.Errorf("Middleware() = %d, want %d", writer.Code, testCase.want)
			}

			if !reflect.DeepEqual(writer.Header(), testCase.wantHeader) {
				t.Errorf("Middleware() = %#v, want %#v", writer.Header(), testCase.wantHeader)
			}
		})
	}
}

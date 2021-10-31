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
	cases := []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -credentials\n    \t[cors] Access-Control-Allow-Credentials {SIMPLE_CREDENTIALS}\n  -expose string\n    \t[cors] Access-Control-Expose-Headers {SIMPLE_EXPOSE}\n  -headers string\n    \t[cors] Access-Control-Allow-Headers {SIMPLE_HEADERS} (default \"Content-Type\")\n  -methods string\n    \t[cors] Access-Control-Allow-Methods {SIMPLE_METHODS} (default \"GET\")\n  -origin string\n    \t[cors] Access-Control-Allow-Origin {SIMPLE_ORIGIN} (default \"*\")\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		intention string
		want      App
	}{
		{
			"simple",
			App{
				origin:      "*",
				headers:     "Content-Type",
				methods:     http.MethodGet,
				credentials: "false",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)

			if result := New(Flags(fs, "")); !reflect.DeepEqual(result, tc.want) {
				t.Errorf("New() = %#v, want %#v", result, tc.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
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
				"Access-Control-Allow-Credentials": []string{"true"},
			},
		},
		{
			"default param",
			App{
				origin:      "*",
				headers:     "Content-Type,Authorization",
				methods:     http.MethodPost,
				exposes:     "X-Total-Count",
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
				"Access-Control-Expose-Headers":    []string{"X-Total-Count"},
				"Access-Control-Allow-Credentials": []string{"false"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			tc.app.Middleware(tc.next).ServeHTTP(writer, tc.request)

			if writer.Code != tc.want {
				t.Errorf("Middleware() = %d, want %d", writer.Code, tc.want)
			}

			if !reflect.DeepEqual(writer.Header(), tc.wantHeader) {
				t.Errorf("Middleware() = %#v, want %#v", writer.Header(), tc.wantHeader)
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	app := App{
		origin:      "*",
		headers:     "Content-Type",
		methods:     http.MethodGet,
		exposes:     "",
		credentials: "true",
	}

	middleware := app.Middleware(nil)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

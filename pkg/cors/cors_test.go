package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		want      map[string]interface{}
	}{
		{
			`default prefix`,
			``,
			map[string]interface{}{
				`origin`:      nil,
				`headers`:     nil,
				`methods`:     nil,
				`exposes`:     nil,
				`credentials`: nil,
			},
		},
		{
			`given prefix`,
			`test`,
			map[string]interface{}{
				`origin`:      nil,
				`headers`:     nil,
				`methods`:     nil,
				`exposes`:     nil,
				`credentials`: nil,
			},
		},
	}

	for _, testCase := range cases {
		if result := Flags(testCase.prefix); len(result) != len(testCase.want) {
			t.Errorf("%v\nFlags(%v) = %v, want %v", testCase.intention, testCase.prefix, result, testCase.want)
		}
	}
}

func Test_ServeHTTP(t *testing.T) {
	var origin = `localhost`
	var headers = `Content-Type,Authorization`
	var methods = `GET,POST`
	var exposes = `X-Total-Count`
	var credentials = true

	var cases = []struct {
		intention string
		config    map[string]interface{}
		want      map[string]string
	}{
		{
			`default values`,
			nil,
			map[string]string{`Access-Control-Allow-Origin`: `*`, `Access-Control-Allow-Headers`: `Content-Type`, `Access-Control-Allow-Methods`: http.MethodGet, `Access-Control-Expose-Headers`: ``, `Access-Control-Allow-Credentials`: `false`},
		},
		{
			`given values`,
			map[string]interface{}{
				`origin`:      &origin,
				`headers`:     &headers,
				`methods`:     &methods,
				`exposes`:     &exposes,
				`credentials`: &credentials,
			},
			map[string]string{`Access-Control-Allow-Origin`: `localhost`, `Access-Control-Allow-Headers`: `Content-Type,Authorization`, `Access-Control-Allow-Methods`: `GET,POST`, `Access-Control-Expose-Headers`: `X-Total-Count`, `Access-Control-Allow-Credentials`: `true`},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRecorder()
		Handler(testCase.config, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(request, nil)

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf("%v\nServeHTTP() -> [%v] = %v, want %v", testCase.intention, key, result, value)
			}
		}
	}
}

func Benchmark_ServeHTTP(b *testing.B) {
	handler := Handler(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(writer, nil)
	}
}

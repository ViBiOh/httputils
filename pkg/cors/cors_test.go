package cors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	origin      = `localhost`
	headers     = `Content-Type,Authorization`
	methods     = fmt.Sprintf(`%s,%s`, http.MethodGet, http.MethodPost)
	exposes     = `X-Total-Count`
	credentials = true
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
	}{
		{
			`should add 5 params to flags`,
			5,
		},
	}

	for _, testCase := range cases {
		if result := Flags(`cors_Test_Flags_empty`); len(result) != testCase.want {
			t.Errorf("%s\nFlags() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_ServeHTTP(t *testing.T) {
	var cases = []struct {
		intention string
		config    map[string]interface{}
		want      map[string]string
	}{
		{
			`default values`,
			Flags(`cors_Test_ServeHTTP_default`),
			map[string]string{
				`Access-Control-Allow-Origin`:      `*`,
				`Access-Control-Allow-Headers`:     `Content-Type`,
				`Access-Control-Allow-Methods`:     http.MethodGet,
				`Access-Control-Expose-Headers`:    ``,
				`Access-Control-Allow-Credentials`: `false`,
			},
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
	handler := Handler(map[string]interface{}{
		`origin`:      &origin,
		`headers`:     &headers,
		`methods`:     &methods,
		`exposes`:     &exposes,
		`credentials`: &credentials,
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(writer, nil)
	}
}

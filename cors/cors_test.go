package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	var tests = []struct {
		want map[string]string
	}{
		{
			map[string]string{`Access-Control-Allow-Origin`: `*`, `Access-Control-Allow-Headers`: `Content-Type`, `Access-Control-Allow-Methods`: `GET`},
		},
	}

	for _, test := range tests {
		request := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(request, nil)

		for key, value := range test.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, result, value)
			}
		}
	}
}

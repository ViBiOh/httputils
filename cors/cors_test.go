package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	var cases = []struct {
		want map[string]string
	}{
		{
			map[string]string{`Access-Control-Allow-Origin`: `*`, `Access-Control-Allow-Headers`: `Content-Type`, `Access-Control-Allow-Methods`: http.MethodGet},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRecorder()
		Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}.ServeHTTP(request, nil)

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, result, value)
			}
		}
	}
}

func BenchmarkServeHTTP(b *testing.B) {
	var testCase = struct {
		want map[string]string
	}{
		map[string]string{`Access-Control-Allow-Origin`: `*`, `Access-Control-Allow-Headers`: `Content-Type`, `Access-Control-Allow-Methods`: http.MethodGet},
	}

	request := httptest.NewRecorder()
	handler := Handler{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(request, nil)

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				b.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, result, value)
			}
		}
	}
}

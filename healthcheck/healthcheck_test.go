package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/request"
)

func Test_Handler(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			`should say ok for GET`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			``,
			http.StatusOK,
		},
		{
			`should say nok for other method`,
			httptest.NewRequest(http.MethodHead, `/`, nil),
			``,
			http.StatusMethodNotAllowed,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		Handler().ServeHTTP(writer, testCase.request)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nHandler(%+v) = %+v, want status %+v", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nHandler(%+v) = %+v, want %+v", testCase.intention, testCase.request, string(result), testCase.want)
		}
	}
}

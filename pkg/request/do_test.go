package request

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_DoAndRead(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/bad` {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Fprint(w, `Hello, test`)
		}
	}))
	defer testServer.Close()

	emptyRequest, _ := http.NewRequest(http.MethodGet, ``, nil)
	bad, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(`%s/bad`, testServer.URL), nil)
	test, _ := http.NewRequest(http.MethodGet, testServer.URL, nil)

	var cases = []struct {
		ctx     context.Context
		request *http.Request
		want    string
		wantErr error
	}{
		{
			nil,
			emptyRequest,
			``,
			errors.New(`Error while processing request: Get : unsupported protocol scheme ""`),
		},
		{
			nil,
			bad,
			``,
			errors.New(`Error status 400`),
		},
		{
			nil,
			test,
			`Hello, test`,
			nil,
		},
		{
			nil,
			test,
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := DoAndRead(testCase.ctx, testCase.request)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if string(result) != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf(`DoAndRead(%v) = (%v, %v), want (%v, %v)`, testCase.request, string(result), err, testCase.want, testCase.wantErr)
		}
	}
}

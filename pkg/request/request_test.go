package request

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testFn() string {
	return "toto"
}

type postStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestDoJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test")
	}))
	defer testServer.Close()

	var cases = []struct {
		ctx     context.Context
		url     string
		body    interface{}
		headers http.Header
		want    string
		wantErr error
	}{
		{
			nil,
			"",
			testFn,
			nil,
			"",
			errors.New("json: unsupported type: func() string"),
		},
		{
			nil,
			"://fail",
			nil,
			nil,
			"",
			errors.New("parse ://fail: missing protocol scheme"),
		},
		{
			nil,
			"",
			nil,
			nil,
			"",
			errors.New("Post : unsupported protocol scheme \"\""),
		},
		{
			nil,
			testServer.URL,
			&postStruct{},
			http.Header{"Authorization": {"admin:password"}},
			"Hello, test",
			nil,
		},
	}

	for _, testCase := range cases {
		result, _, _, err := DoJSON(testCase.ctx, testCase.url, testCase.body, testCase.headers, http.MethodPost)

		failed := false
		var content []byte

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != nil {
			content, _ = ReadBody(result)

			if string(content) != testCase.want {
				failed = true
			}
		}

		if failed {
			t.Errorf("PostJSONBody(%#v, %#v, '') = (%s, %#v), want (%s, %#v)", testCase.url, testCase.body, string(content), err, testCase.want, testCase.wantErr)
		}
	}
}

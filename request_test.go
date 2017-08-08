package httputils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddAuthorization(t *testing.T) {
	var tests = []struct {
		authorization string
	}{
		{
			``,
		},
		{
			`admin`,
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(`GET`, `http://localhost`, nil)
		addAuthorization(request, test.authorization)

		if result := strings.Join(request.Header[`Authorization`], ``); result != test.authorization {
			t.Errorf(`addAuthorization(%v) = %v, want %v`, test.authorization, result, test.authorization)
		}
	}
}

func TestGetBasicAuth(t *testing.T) {
	var tests = []struct {
		username string
		password string
		want     string
	}{
		{
			``,
			``,
			`Basic Og==`,
		},
		{
			`admin`,
			`password`,
			`Basic YWRtaW46cGFzc3dvcmQ=`,
		},
	}

	for _, test := range tests {
		if result := GetBasicAuth(test.username, test.password); result != test.want {
			t.Errorf(`GetBasicAuth(%v, %v) = %v, want %v`, test.username, test.password, result, test.want)
		}
	}
}

func TestReadBody(t *testing.T) {
	var tests = []struct {
		body    io.ReadCloser
		want    string
		wantErr error
	}{
		{
			ioutil.NopCloser(bytes.NewBuffer([]byte(`Content from buffer`))),
			`Content from buffer`,
			nil,
		},
	}

	var failed bool

	for _, test := range tests {
		result, err := ReadBody(test.body)

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		} else if string(result) != test.want {
			failed = true
		}

		if failed {
			t.Errorf(`ReadBody(%v) = (%v, %v), want (%v, %v)`, test.body, result, err, test.want, test.wantErr)
		}
	}
}

func TestGetBody(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `Hello, test`)
	}))
	defer testServer.Close()

	var tests = []struct {
		url     string
		want    string
		wantErr error
	}{
		{
			testServer.URL,
			`Hello, test
`,
			nil,
		},
	}

	var failed bool

	for _, test := range tests {
		result, err := GetBody(test.url, ``)

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		} else if string(result) != test.want {
			failed = true
		}

		if failed {
			t.Errorf(`GetBody(%v, '') = (%s, %v), want (%s, %v)`, test.url, result, err, test.want, test.wantErr)
		}
	}
}

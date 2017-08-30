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

type postStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestDoAndRead(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/bad` {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Fprint(w, `Hello, test`)
		}
	}))
	defer testServer.Close()

	emptyRequest, _ := http.NewRequest(http.MethodGet, ``, nil)
	bad, _ := http.NewRequest(http.MethodGet, testServer.URL+`/bad`, nil)
	test, _ := http.NewRequest(http.MethodGet, testServer.URL, nil)

	var cases = []struct {
		request       *http.Request
		skipTLSVerify bool
		want          string
		wantErr       error
	}{
		{
			emptyRequest,
			false,
			``,
			fmt.Errorf(`Error while sending data: Get : unsupported protocol scheme ""`),
		},
		{
			bad,
			false,
			``,
			fmt.Errorf(`Error status 400: `),
		},
		{
			test,
			false,
			`Hello, test`,
			nil,
		},
		{
			test,
			true,
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := doAndRead(testCase.request, testCase.skipTLSVerify)

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
			t.Errorf(`doAndRead(%v) = (%v, %v), want (%v, %v)`, testCase.request, string(result), err, testCase.want, testCase.wantErr)
		}
	}
}

func TestAddAuthorization(t *testing.T) {
	var cases = []struct {
		authorization string
	}{
		{
			``,
		},
		{
			`admin`,
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRequest(http.MethodGet, `http://localhost`, nil)
		addAuthorization(request, testCase.authorization)

		if result := strings.Join(request.Header[`Authorization`], ``); result != testCase.authorization {
			t.Errorf(`addAuthorization(%v) = %v, want %v`, testCase.authorization, result, testCase.authorization)
		}
	}
}

func TestGetBasicAuth(t *testing.T) {
	var cases = []struct {
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

	for _, testCase := range cases {
		if result := GetBasicAuth(testCase.username, testCase.password); result != testCase.want {
			t.Errorf(`GetBasicAuth(%v, %v) = %v, want %v`, testCase.username, testCase.password, result, testCase.want)
		}
	}
}

func TestReadBody(t *testing.T) {
	var cases = []struct {
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

	for _, testCase := range cases {
		result, err := ReadBody(testCase.body)

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
			t.Errorf(`ReadBody(%v) = (%v, %v), want (%v, %v)`, testCase.body, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func TestGetBody(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `Hello, test`)
	}))
	defer testServer.Close()

	var cases = []struct {
		url     string
		want    string
		wantErr error
	}{
		{
			`://fail`,
			``,
			fmt.Errorf(`Error while creating request: parse ://fail: missing protocol scheme`),
		},
		{
			testServer.URL,
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := GetBody(testCase.url, ``, false)

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
			t.Errorf(`GetBody(%v, '') = (%s, %v), want (%s, %v)`, testCase.url, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func TestPostJSONBody(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `Hello, test`)
	}))
	defer testServer.Close()

	var cases = []struct {
		url     string
		body    interface{}
		want    string
		wantErr error
	}{
		{
			``,
			testFn,
			``,
			fmt.Errorf(`Error while marshalling body: json: unsupported type: func() string`),
		},
		{
			`://fail`,
			nil,
			``,
			fmt.Errorf(`Error while creating request: parse ://fail: missing protocol scheme`),
		},
		{
			``,
			nil,
			``,
			fmt.Errorf(`Error while sending data: Post : unsupported protocol scheme ""`),
		},
		{
			testServer.URL,
			&postStruct{},
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := PostJSONBody(testCase.url, testCase.body, ``, false)

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
			t.Errorf(`PostJSONBody(%v, %v, '') = (%s, %v), want (%s, %v)`, testCase.url, testCase.body, result, err, testCase.want, testCase.wantErr)
		}
	}
}

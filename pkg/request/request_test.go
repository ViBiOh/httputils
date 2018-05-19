package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testFn() string {
	return `toto`
}

type postStruct struct {
	id     string
	Active bool
	Amount float64
}

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
		request *http.Request
		want    string
		wantErr error
	}{
		{
			emptyRequest,
			``,
			errors.New(`Error while processing request: Get : unsupported protocol scheme ""`),
		},
		{
			bad,
			``,
			errors.New(`Error status 400`),
		},
		{
			test,
			`Hello, test`,
			nil,
		},
		{
			test,
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := doAndRead(testCase.request)

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

func Test_GetBasicAuth(t *testing.T) {
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

func Test_ReadBody(t *testing.T) {
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

func Test_DoJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `Hello, test`)
	}))
	defer testServer.Close()

	var cases = []struct {
		url     string
		body    interface{}
		headers http.Header
		want    string
		wantErr error
	}{
		{
			``,
			testFn,
			nil,
			``,
			errors.New(`Error while marshalling body: json: unsupported type: func() string`),
		},
		{
			`://fail`,
			nil,
			nil,
			``,
			errors.New(`Error while creating request: parse ://fail: missing protocol scheme`),
		},
		{
			``,
			nil,
			nil,
			``,
			errors.New(`Error while processing request: Post : unsupported protocol scheme ""`),
		},
		{
			testServer.URL,
			&postStruct{},
			http.Header{`Authorization`: []string{`admin:password`}},
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := DoJSON(testCase.url, testCase.body, testCase.headers, http.MethodPost)

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

func Test_SetIP(t *testing.T) {
	var cases = []struct {
		intention string
		request   *http.Request
		ip        string
	}{
		{
			`should set header with given string`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			`test`,
		},
	}

	for _, testCase := range cases {
		if SetIP(testCase.request, testCase.ip); testCase.request.Header.Get(ForwardedForHeader) != testCase.ip {
			t.Errorf("%s\nSetIP(%+v, %+v) = %+v, want %+v", testCase.intention, testCase.request, testCase.ip, testCase.request.Header.Get(ForwardedForHeader), testCase.ip)
		}
	}
}

func Test_GetIP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/`, nil)
	request.RemoteAddr = `localhost`

	requestWithProxy := httptest.NewRequest(http.MethodGet, `/`, nil)
	requestWithProxy.RemoteAddr = `localhost`
	requestWithProxy.Header.Set(ForwardedForHeader, `proxy`)

	var cases = []struct {
		r    *http.Request
		want string
	}{
		{
			request,
			`localhost`,
		},
		{
			requestWithProxy,
			`proxy`,
		},
	}

	for _, testCase := range cases {
		if result := GetIP(testCase.r); result != testCase.want {
			t.Errorf(`GetIP(%v) = %v, want %v`, testCase.r, result, testCase.want)
		}
	}
}

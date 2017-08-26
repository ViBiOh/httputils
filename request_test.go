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

	var tests = []struct {
		request *http.Request
		want    string
		wantErr error
	}{
		{
			emptyRequest,
			``,
			fmt.Errorf(`Error while sending data: Get : unsupported protocol scheme ""`),
		},
		{
			bad,
			``,
			fmt.Errorf(`Error status 400: `),
		},
		{
			test,
			`Hello, test`,
			nil,
		},
	}

	var failed bool

	for _, test := range tests {
		result, err := doAndRead(test.request, false)

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
			t.Errorf(`doAndRead(%v) = (%v, %v), want (%v, %v)`, test.request, string(result), err, test.want, test.wantErr)
		}
	}
}

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
		request := httptest.NewRequest(http.MethodGet, `http://localhost`, nil)
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
		fmt.Fprint(w, `Hello, test`)
	}))
	defer testServer.Close()

	var tests = []struct {
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

	for _, test := range tests {
		result, err := GetBody(test.url, ``, false)

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

func TestPostJSONBody(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `Hello, test`)
	}))
	defer testServer.Close()

	var tests = []struct {
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

	for _, test := range tests {
		result, err := PostJSONBody(test.url, test.body, ``, false)

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
			t.Errorf(`PostJSONBody(%v, %v, '') = (%s, %v), want (%s, %v)`, test.url, test.body, result, err, test.want, test.wantErr)
		}
	}
}

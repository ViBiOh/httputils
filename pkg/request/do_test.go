package request

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestDo(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		if r.URL.Path == "/invalid" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid"))
			return
		}

		if r.URL.Path == "/internalError" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "https://vibioh.fr")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		}
	}))
	defer testServer.Close()

	simple, _ := http.NewRequest(http.MethodGet, testServer.URL+"/simple", nil)
	invalid, _ := http.NewRequest(http.MethodGet, testServer.URL+"/invalid", nil)
	internalError, _ := http.NewRequest(http.MethodGet, testServer.URL+"/internalError", nil)
	redirect, _ := http.NewRequest(http.MethodGet, testServer.URL+"/redirect", nil)

	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
		wantErr    error
	}{
		{
			"simple",
			simple,
			"valid",
			http.StatusOK,
			nil,
		},
		{
			"invalid",
			invalid,
			"",
			http.StatusBadRequest,
			errors.New("HTTP/400\ninvalid"),
		},
		{
			"internalError",
			internalError,
			"",
			http.StatusInternalServerError,
			errors.New("HTTP/500"),
		},
		{
			"redirect",
			redirect,
			"",
			http.StatusPermanentRedirect,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := Do(testCase.request)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			} else if resp.StatusCode != testCase.wantStatus {
				failed = true
			}

			if failed {
				t.Errorf("Do() = (%s, %d, %#v), want (%s, %d, %#v)", result, resp.StatusCode, err, testCase.want, testCase.wantStatus, testCase.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		ctx       context.Context
		url       string
		want      string
		wantErr   error
	}{
		{
			"simple",
			context.Background(),
			testServer.URL + "/simple",
			"valid",
			nil,
		},
		{
			"invalid request",
			nil,
			"",
			"",
			errors.New("net/http: nil Context"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := Get(testCase.ctx, testCase.url, nil)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%s, %#v), want (%s, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestPost(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" && r.Method == http.MethodPost && r.FormValue("first") == "test" && r.FormValue("second") == "param" && r.Header.Get(ContentTypeHeader) == "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		ctx       context.Context
		url       string
		data      url.Values
		want      string
		wantErr   error
	}{
		{
			"simple",
			context.Background(),
			testServer.URL + "/simple",
			url.Values{
				"first":  []string{"test"},
				"second": []string{"param"},
			},
			"valid",
			nil,
		},
		{
			"invalid request",
			nil,
			"",
			nil,
			"",
			errors.New("net/http: nil Context"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := Post(testCase.ctx, testCase.url, testCase.data, nil)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("Post() = (%s, %#v), want (%s, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := ReadBodyRequest(r)

		if r.URL.Path == "/simple" && r.Method == http.MethodPost && string(payload) == "{\"Active\":true,\"Amount\":12.34}" && r.Header.Get(ContentTypeHeader) == "application/json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		ctx       context.Context
		url       string
		data      interface{}
		want      string
		wantErr   error
	}{
		{
			"simple",
			context.Background(),
			testServer.URL + "/simple",
			testStruct{id: "Test", Active: true, Amount: 12.34},
			"valid",
			nil,
		},
		{
			"invalid request",
			nil,
			"",
			nil,
			"",
			errors.New("net/http: nil Context"),
		},
		{
			"invalid marshall",
			context.Background(),
			"",
			func() string {
				return "test"
			},
			"",
			errors.New("json: unsupported type: func() string"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := PostJSON(testCase.ctx, testCase.url, testCase.data, nil)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("PostJSON() = (%s, %#v), want (%s, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

package request

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestSend(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" {
			w.WriteHeader(http.StatusOK)

			payload, _ := ReadBodyRequest(r)

			if r.Method == http.MethodGet {
				w.Write([]byte("it works!"))
			} else if r.Method == http.MethodPost && string(payload) == "posted" {
				w.Write([]byte("it posts!"))
			} else if r.Method == http.MethodPut && string(payload) == "puted" {
				w.Write([]byte("it puts!"))
			} else if r.Method == http.MethodPatch && string(payload) == "patched" {
				w.Write([]byte("it patches!"))
			} else if r.Method == http.MethodDelete {
				w.Write([]byte("it deletes!"))
			}

			return
		} else if r.URL.Path == "/protected" {
			username, password, ok := r.BasicAuth()
			if ok && username == "admin" && password == "secret" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("connected!"))

				return
			}
		} else if r.URL.Path == "/accept" {
			if r.Header.Get("Accept") == "text/plain" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("text me!"))

				return
			}
		} else if r.URL.Path == "/explain" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing id"))

			return
		} else if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "/simple")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		} else if r.URL.Path == "/client" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		request   *Request
		ctx       context.Context
		payload   io.Reader
		want      string
		wantErr   error
	}{
		{
			"simple get",
			New().Get(testServer.URL + "/simple"),
			context.Background(),
			nil,
			"it works!",
			nil,
		},
		{
			"simple post",
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			strings.NewReader("posted"),
			"it posts!",
			nil,
		},
		{
			"simple put",
			New().Put(testServer.URL + "/simple"),
			context.Background(),
			strings.NewReader("puted"),
			"it puts!",
			nil,
		},
		{
			"simple patch",
			New().Patch(testServer.URL + "/simple"),
			context.Background(),
			strings.NewReader("patched"),
			"it patches!",
			nil,
		},
		{
			"simple delete",
			New().Delete(testServer.URL + "/simple"),
			context.Background(),
			nil,
			"it deletes!",
			nil,
		},
		{
			"with auth",
			New().Get(testServer.URL+"/protected").BasicAuth("admin", "secret"),
			context.Background(),
			nil,
			"connected!",
			nil,
		},
		{
			"with header",
			New().Get(testServer.URL+"/accept").Header("Accept", "text/plain"),
			context.Background(),
			nil,
			"text me!",
			nil,
		},
		{
			"with client",
			New().Get(testServer.URL + "/client").WithClient(http.Client{}),
			context.Background(),
			nil,
			"",
			nil,
		},
		{
			"invalid request",
			New().Get(testServer.URL + "/invalid"),
			nil,
			nil,
			"",
			errors.New("net/http: nil Context"),
		},
		{
			"invalid status code",
			New().Get(testServer.URL + "/invalid"),
			context.Background(),
			nil,
			"",
			errors.New("HTTP/500"),
		},
		{
			"invalid status code with payload",
			New().Get(testServer.URL + "/explain"),
			context.Background(),
			nil,
			"",
			errors.New("HTTP/400\nmissing id"),
		},
		{
			"don't redirect",
			New().Get(testServer.URL + "/redirect"),
			context.Background(),
			nil,
			"",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := testCase.request.Send(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

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
				t.Errorf("Send() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestForm(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" && r.Method == http.MethodPost && r.FormValue("first") == "test" && r.FormValue("second") == "param" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		request   *Request
		ctx       context.Context
		payload   url.Values
		want      string
		wantErr   error
	}{
		{
			"simple",
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			url.Values{
				"first":  []string{"test"},
				"second": []string{"param"},
			},
			"valid",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := testCase.request.Form(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

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
				t.Errorf("Send() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := ReadBodyRequest(r)

		if r.URL.Path == "/simple" && r.Method == http.MethodPost && string(payload) == "{\"Active\":true,\"Amount\":12.34}" && r.Header.Get("Content-Type") == "application/json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		request   *Request
		ctx       context.Context
		payload   interface{}
		want      string
		wantErr   error
	}{
		{
			"simple",
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			testStruct{id: "Test", Active: true, Amount: 12.34},
			"valid",
			nil,
		},
		{
			"invalid",
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			func() {},
			"",
			errors.New("json: unsupported type: func()"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			resp, err := testCase.request.JSON(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

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
				t.Errorf("Send() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

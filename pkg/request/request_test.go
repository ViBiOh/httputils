package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/golang/mock/gomock"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func safeWrite(writer io.Writer, content []byte) {
	if _, err := writer.Write(content); err != nil {
		fmt.Println(err)
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		intention string
		instance  Request
		want      string
	}{
		{
			"simple",
			New(),
			"GET",
		},
		{
			"basic auth",
			New().Post("http://localhost").BasicAuth("admin", "password").ContentType("text/plain"),
			"POST http://localhost, BasicAuth with user `%s`admin, Header Content-Type: `text/plain`",
		},
		{
			"signature auth",
			New().Post("http://localhost").WithSignatureAuthorization("secret", []byte("password")),
			"POST http://localhost, SignatureAuthorization with key `secret`",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.String(); got != tc.want {
				t.Errorf("String() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	cases := []struct {
		intention string
		instance  Request
		want      bool
	}{
		{
			"empty",
			New(),
			true,
		},
		{
			"simple",
			New().Get("/"),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.IsZero(); got != tc.want {
				t.Errorf("IsZero() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	type args struct {
		path string
	}

	cases := []struct {
		intention string
		instance  Request
		args      args
		want      Request
	}{
		{
			"empty",
			New().Get("http://localhost"),
			args{
				path: "",
			},
			New().Get("http://localhost"),
		},
		{
			"no prefix",
			New().Get("http://localhost"),
			args{
				path: "hello",
			},
			New().Get("http://localhost/hello"),
		},
		{
			"trailing slash url",
			New().Get("http://localhost/"),
			args{
				path: "hello",
			},
			New().Get("http://localhost/hello"),
		},
		{
			"prefix path",
			New().Get("http://localhost"),
			args{
				path: "/hello",
			},
			New().Get("http://localhost/hello"),
		},
		{
			"full slash",
			New().Get("http://localhost/"),
			args{
				path: "/hello",
			},
			New().Get("http://localhost/hello"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.Path(tc.args.path); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Path() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestSend(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" {
			w.WriteHeader(http.StatusOK)

			payload, _ := ReadBodyRequest(r)

			if r.Method == http.MethodGet {
				safeWrite(w, []byte("it works!"))
			} else if r.Method == http.MethodPost && string(payload) == "posted" {
				safeWrite(w, []byte("it posts!"))
			} else if r.Method == http.MethodPut && string(payload) == "puted" {
				safeWrite(w, []byte("it puts!"))
			} else if r.Method == http.MethodPatch && string(payload) == "patched" {
				safeWrite(w, []byte("it patches!"))
			} else if r.Method == http.MethodDelete {
				safeWrite(w, []byte("it deletes!"))
			}

			return
		} else if r.URL.Path == "/protected" {
			username, password, ok := r.BasicAuth()
			if ok && username == "admin" && password == "secret" {
				w.WriteHeader(http.StatusOK)
				safeWrite(w, []byte("connected!"))

				return
			}
		} else if r.URL.Path == "/accept" {
			if r.Header.Get("Accept") == "text/plain" {
				w.WriteHeader(http.StatusOK)
				safeWrite(w, []byte("text me!"))

				return
			}
		} else if r.URL.Path == "/explain" {
			w.WriteHeader(http.StatusBadRequest)
			safeWrite(w, []byte("missing id"))

			return
		} else if r.URL.Path == "/long_explain" {
			w.WriteHeader(http.StatusBadRequest)
			safeWrite(w, []byte(`Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`))
			return
		} else if r.URL.Path == "/redirect" {
			w.Header().Add("Location", "/simple")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		} else if r.URL.Path == "/client" {
			w.WriteHeader(http.StatusNoContent)
			return
		} else if r.URL.Path == "/timeout" {
			time.Sleep(time.Second * 2)
			w.WriteHeader(http.StatusNoContent)
			return
		} else if r.URL.Path == "/signed" {
			if ok, err := ValidateSignature(r, []byte(`secret`)); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				safeWrite(w, []byte(err.Error()))
			} else if ok {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
				safeWrite(w, []byte("signature doesn't match"))
			}
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	defaultHTTPClient = CreateClient(time.Second, NoRedirection)

	cases := []struct {
		intention string
		request   Request
		ctx       context.Context
		payload   io.ReadCloser
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
			io.NopCloser(strings.NewReader("posted")),
			"it posts!",
			nil,
		},
		{
			"simple put",
			New().Put(testServer.URL + "/simple"),
			context.Background(),
			io.NopCloser(strings.NewReader("puted")),
			"it puts!",
			nil,
		},
		{
			"simple patch",
			New().Patch(testServer.URL + "/simple"),
			context.Background(),
			io.NopCloser(strings.NewReader("patched")),
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
			New().Get(testServer.URL + "/client").WithClient(&http.Client{}),
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
			errors.New("HTTP/400"),
		},
		{
			"invalid status code with long payload",
			New().Get(testServer.URL + "/long_explain"),
			context.Background(),
			nil,
			"",
			errors.New("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing e"),
		},
		{
			"don't redirect",
			New().Get(testServer.URL + "/redirect"),
			context.Background(),
			nil,
			"",
			nil,
		},
		{
			"timeout",
			New().Get(testServer.URL + "/timeout"),
			context.Background(),
			nil,
			"",
			errors.New("context deadline exceeded (Client.Timeout exceeded while awaiting headers)"),
		},
		{
			"signed",
			New().Post(testServer.URL+"/signed").WithSignatureAuthorization("httputils", []byte(`secret`)),
			context.Background(),
			io.NopCloser(strings.NewReader(`It works!`)),
			"",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			resp, err := tc.request.Send(tc.ctx, tc.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if err == nil && tc.wantErr != nil {
				failed = true
			} else if err != nil && tc.wantErr == nil {
				failed = true
			} else if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestForm(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" && r.Method == http.MethodPost && r.FormValue("first") == "test" && r.FormValue("second") == "param" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusOK)
			safeWrite(w, []byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	cases := []struct {
		intention string
		request   Request
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			resp, err := tc.request.Form(tc.ctx, tc.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if err == nil && tc.wantErr != nil {
				failed = true
			} else if err != nil && tc.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != tc.wantErr.Error() {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := ReadBodyRequest(r)

		if r.URL.Path == "/simple" && r.Method == http.MethodPost && string(payload) == "{\"Active\":true,\"Amount\":12.34}\n" && r.Header.Get("Content-Type") == "application/json" {
			w.WriteHeader(http.StatusOK)
			safeWrite(w, []byte("valid"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	cases := []struct {
		intention string
		request   Request
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			resp, err := tc.request.JSON(tc.ctx, tc.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			if err == nil && tc.wantErr != nil {
				failed = true
			} else if err != nil && tc.wantErr == nil {
				failed = true
			} else if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestDiscardBody(t *testing.T) {
	cases := []struct {
		intention string
		wantErr   error
	}{
		{
			"read error",
			errors.New("read error"),
		},
		{
			"close error",
			errors.New("close error"),
		},
		{
			"valid",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockReadCloser := mocks.NewReadCloser(ctrl)

			body := mockReadCloser

			switch tc.intention {
			case "read error":
				mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
				mockReadCloser.EXPECT().Close().Return(nil)
			case "close error":
				mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
				mockReadCloser.EXPECT().Close().Return(errors.New("close error"))
			case "valid":
				mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
				mockReadCloser.EXPECT().Close().Return(nil)
			}

			gotErr := DiscardBody(body)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("DiscardBody() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

func BenchmarkJSON(b *testing.B) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := ReadBodyRequest(r); err != nil {
			b.Error(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	ctx := context.Background()
	req := New().Post(testServer.URL + "/simple")
	payload := testStruct{id: "Test", Active: true, Amount: 12.34}

	for i := 0; i < b.N; i++ {
		if _, err := req.JSON(ctx, &payload); err != nil {
			b.Error(err)
		}
	}
}

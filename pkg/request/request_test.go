package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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
	t.Parallel()

	cases := map[string]struct {
		instance Request
		want     string
	}{
		"simple": {
			New().ContentLength(8000),
			"GET, ContentLength: 8000",
		},
		"basic auth": {
			Post("http://localhost").BasicAuth("admin", "password").ContentType("text/plain"),
			"POST http://localhost, BasicAuth with user `%s`admin, Header Content-Type: `text/plain`",
		},
		"signature auth": {
			Post("http://localhost").WithSignatureAuthorization("secret", []byte("password")),
			"POST http://localhost, SignatureAuthorization with key `secret`",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.String(); got != testCase.want {
				t.Errorf("String() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance Request
		want     bool
	}{
		"empty": {
			New(),
			true,
		},
		"simple": {
			New().Get("/"),
			false,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.IsZero(); got != testCase.want {
				t.Errorf("IsZero() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()

	type args struct {
		path string
		args []any
	}

	cases := map[string]struct {
		instance Request
		args     args
		want     Request
	}{
		"empty": {
			Get("http://localhost"),
			args{
				path: "",
			},
			Get("http://localhost"),
		},
		"no prefix": {
			Put("http://localhost"),
			args{
				path: "hello",
			},
			Put("http://localhost/hello"),
		},
		"trailing slash url": {
			Post("http://localhost/"),
			args{
				path: "hello",
			},
			Post("http://localhost/hello"),
		},
		"prefix path": {
			Patch("http://localhost"),
			args{
				path: "/hello",
			},
			Patch("http://localhost/hello"),
		},
		"full slash": {
			Delete("http://localhost/"),
			args{
				path: "/hello",
			},
			Delete("http://localhost/hello"),
		},
		"sprintf slash": {
			Delete("http://localhost/"),
			args{
				path: "/hello/%s",
				args: []any{"world"},
			},
			Delete("http://localhost/hello/world"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.Path(testCase.args.path, testCase.args.args...); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("Path() = %#v, want %#v", got, testCase.want)
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
			w.Header().Add("X-Test-Value", "Value with placehodler %d like this %s")
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

	cases := map[string]struct {
		request Request
		ctx     context.Context
		payload io.ReadCloser
		want    string
		wantErr error
	}{
		"simple get": {
			New().Get(testServer.URL + "/simple"),
			context.Background(),
			nil,
			"it works!",
			nil,
		},
		"simple post": {
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			io.NopCloser(strings.NewReader("posted")),
			"it posts!",
			nil,
		},
		"simple put": {
			New().Put(testServer.URL + "/simple"),
			context.Background(),
			io.NopCloser(strings.NewReader("puted")),
			"it puts!",
			nil,
		},
		"simple patch": {
			New().Patch(testServer.URL + "/simple"),
			context.Background(),
			io.NopCloser(strings.NewReader("patched")),
			"it patches!",
			nil,
		},
		"simple delete": {
			New().Delete(testServer.URL + "/simple"),
			context.Background(),
			nil,
			"it deletes!",
			nil,
		},
		"with auth": {
			New().Get(testServer.URL+"/protected").BasicAuth("admin", "secret"),
			context.Background(),
			nil,
			"connected!",
			nil,
		},
		"with header": {
			New().Get(testServer.URL+"/accept").Header("Accept", "text/plain"),
			context.Background(),
			nil,
			"text me!",
			nil,
		},
		"with client": {
			New().Get(testServer.URL + "/client").WithClient(&http.Client{}),
			context.Background(),
			nil,
			"",
			nil,
		},
		"invalid request": {
			New().Get(testServer.URL + "/invalid"),
			nil,
			nil,
			"",
			errors.New("net/http: nil Context"),
		},
		"invalid status code": {
			New().Get(testServer.URL + "/invalid"),
			context.Background(),
			nil,
			"",
			errors.New("HTTP/500"),
		},
		"invalid status code with payload": {
			New().Get(testServer.URL + "/explain"),
			context.Background(),
			nil,
			"",
			errors.New("HTTP/400"),
		},
		"invalid status code with long payload": {
			New().Get(testServer.URL + "/long_explain"),
			context.Background(),
			nil,
			"",
			errors.New("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing e"),
		},
		"don't redirect": {
			New().Get(testServer.URL + "/redirect"),
			context.Background(),
			nil,
			"",
			nil,
		},
		"timeout": {
			New().Get(testServer.URL + "/timeout"),
			context.Background(),
			nil,
			"",
			errors.New("context deadline exceeded (Client.Timeout exceeded while awaiting headers)"),
		},
		"signed": {
			New().Post(testServer.URL+"/signed").WithSignatureAuthorization("httputils", []byte(`secret`)),
			context.Background(),
			io.NopCloser(strings.NewReader(`It works!`)),
			"",
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			resp, err := testCase.request.Send(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			switch {
			case
				testCase.wantErr == nil && err != nil,
				testCase.wantErr != nil && err == nil,
				testCase.wantErr != nil && err != nil && !strings.Contains(err.Error(), testCase.wantErr.Error()),
				string(result) != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestForm(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" && r.Method == http.MethodPost && r.FormValue("first") == "test" && r.FormValue("second") == "param" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusOK)
			safeWrite(w, []byte("valid"))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	cases := map[string]struct {
		request Request
		ctx     context.Context
		payload url.Values
		want    string
		wantErr error
	}{
		"simple": {
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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
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
				t.Errorf("Form() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestMultipart(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/data") {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
		safeWrite(w, []byte(r.FormValue("hello")))
	}))
	defer testServer.Close()

	cases := map[string]struct {
		request Request
		ctx     context.Context
		feed    func(mw *multipart.Writer) error
		want    string
		wantErr error
	}{
		"simple": {
			New().Post(testServer.URL),
			context.Background(),
			func(mw *multipart.Writer) error {
				return mw.WriteField("hello", "world")
			},
			"world",
			nil,
		},
		"with file": {
			New().Post(testServer.URL),
			context.Background(),
			func(mw *multipart.Writer) error {
				header := textproto.MIMEHeader{}
				header.Set("Content-Disposition", `form-data; name="hello"`)
				header.Set("Content-Type", "application/json")

				writer, err := mw.CreatePart(header)
				if err != nil {
					return err
				}

				return json.NewEncoder(writer).Encode(map[string]string{"hello": "world"})
			},
			`{"hello":"world"}
`,
			nil,
		},
		"feed error": {
			New().Post(testServer.URL),
			context.Background(),
			func(mw *multipart.Writer) error {
				return errors.New("failed")
			},
			``,
			errors.New("failed"),
		},
		"server error": {
			New().Get(testServer.URL),
			context.Background(),
			func(mw *multipart.Writer) error {
				return errors.New("failed")
			},
			``,
			errors.New("HTTP/400"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			resp, err := testCase.request.Multipart(testCase.ctx, testCase.feed)
			result, _ := ReadBodyResponse(resp)

			failed := false

			switch {
			case
				testCase.wantErr == nil && err != nil,
				testCase.wantErr != nil && err == nil,
				testCase.wantErr != nil && err != nil && !strings.HasPrefix(err.Error(), testCase.wantErr.Error()),
				string(result) != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("Multipart() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := ReadBodyRequest(r)

		if r.URL.Path == "/simple" && r.Method == http.MethodPost && string(payload) == `{"Active":true,"Amount":12.34}` && r.Header.Get("Content-Type") == "application/json" {
			w.WriteHeader(http.StatusOK)
			safeWrite(w, []byte("valid"))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	cases := map[string]struct {
		request Request
		ctx     context.Context
		payload any
		want    string
		wantErr error
	}{
		"simple": {
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			testStruct{id: "Test", Active: true, Amount: 12.34},
			"valid",
			nil,
		},
		"invalid": {
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			func() {},
			"",
			errors.New("json: unsupported type: func()"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			resp, err := testCase.request.JSON(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			switch {
			case
				testCase.wantErr == nil && err != nil,
				testCase.wantErr != nil && err == nil,
				testCase.wantErr != nil && err != nil && !strings.Contains(err.Error(), testCase.wantErr.Error()),
				string(result) != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestStreamJSON(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := ReadBodyRequest(r)

		if r.URL.Path == "/simple" && r.Method == http.MethodPost && string(payload) == `{"Active":true,"Amount":12.34}`+"\n" && r.Header.Get("Content-Type") == "application/json" {
			w.WriteHeader(http.StatusOK)
			safeWrite(w, []byte("valid"))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	cases := map[string]struct {
		request Request
		ctx     context.Context
		payload any
		want    string
		wantErr error
	}{
		"simple": {
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			testStruct{id: "Test", Active: true, Amount: 12.34},
			"valid",
			nil,
		},
		"invalid": {
			New().Post(testServer.URL + "/simple"),
			context.Background(),
			func() {},
			"",
			errors.New("json: unsupported type: func()"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			resp, err := testCase.request.StreamJSON(testCase.ctx, testCase.payload)
			result, _ := ReadBodyResponse(resp)

			failed := false

			switch {
			case
				testCase.wantErr == nil && err != nil,
				testCase.wantErr != nil && err == nil,
				testCase.wantErr != nil && err != nil && !strings.Contains(err.Error(), testCase.wantErr.Error()),
				string(result) != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("Send() = (`%s`,`%s`), want (`%s`,`%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestDiscardBody(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"read error": {
			errors.New("read error"),
		},
		"close error": {
			errors.New("close error"),
		},
		"valid": {
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockReadCloser := mocks.NewReadCloser(ctrl)

			body := mockReadCloser

			switch intention {
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

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("DiscardBody() = `%s`, want `%s`", gotErr, testCase.wantErr)
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

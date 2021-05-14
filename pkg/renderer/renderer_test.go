package renderer

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -publicURL string\n    \tPublic URL {SIMPLE_PUBLIC_URL} (default \"http://localhost\")\n  -title string\n    \tApplication title {SIMPLE_TITLE} (default \"App\")\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestIsStaticRootPaths(t *testing.T) {
	type args struct {
		requestPath string
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
	}{
		{
			"empty",
			args{
				requestPath: "/",
			},
			false,
		},
		{
			"robots",
			args{
				requestPath: "/robots.txt",
			},
			true,
		},
		{
			"sitemap",
			args{
				requestPath: "/sitemap.xml",
			},
			true,
		},
		{
			"subpath",
			args{
				requestPath: "/test/sitemap.xml",
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := isStaticRootPaths(tc.args.requestPath); got != tc.want {
				t.Errorf("isStaticRootPaths() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestFeedContent(t *testing.T) {
	type args struct {
		content map[string]interface{}
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      map[string]interface{}
	}{
		{
			"empty",
			app{},
			args{
				content: nil,
			},
			make(map[string]interface{}),
		},
		{
			"merge",
			app{
				content: map[string]interface{}{
					"Version": "test",
				},
			},
			args{
				content: map[string]interface{}{
					"Name": "Hello World",
				},
			},
			map[string]interface{}{
				"Version": "test",
				"Name":    "Hello World",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.feedContent(tc.args.content); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("feedContent() = %v, want %v", tc.args.content, tc.want)
			}
		})
	}
}

//go:embed templates static
var content embed.FS

func TestHandler(t *testing.T) {
	publicURL := "http://localhost"
	title := "Golang Test"

	configuredApp, err := New(Config{
		publicURL: &publicURL,
		title:     &title,
	}, content, template.FuncMap{})
	if err != nil {
		t.Error(err)
	}

	var cases = []struct {
		intention    string
		instance     App
		request      *http.Request
		templateFunc TemplateFunc
		want         string
		wantStatus   int
		wantHeader   http.Header
	}{
		{
			"empty app",
			app{},
			httptest.NewRequest(http.MethodGet, "/", nil),
			nil,
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
			http.Header{},
		},
		{
			"favicon",
			configuredApp,
			httptest.NewRequest(http.MethodGet, "/favicon/manifest.json", nil),
			nil,
			"{}\n",
			http.StatusOK,
			http.Header{},
		},
		{
			"svg",
			configuredApp,
			httptest.NewRequest(http.MethodGet, "/svg/test?fill=black", nil),
			nil,
			"color=black",
			http.StatusOK,
			http.Header{},
		},
		{
			"svg not found",
			configuredApp,
			httptest.NewRequest(http.MethodGet, "/svg/unknown?fill=black", nil),
			nil,
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
			http.Header{},
		},
		{
			"html",
			configuredApp,
			httptest.NewRequest(http.MethodGet, "/", nil),
			func(_ *http.Request) (string, int, map[string]interface{}, error) {
				return "public", http.StatusCreated, nil, nil
			},
			`<!doctype html><html lang=en><meta charset=utf-8><title>Golang Test</title><h1>Hello !</h1>`,
			http.StatusCreated,
			http.Header{},
		},
		{
			"message",
			configuredApp,
			httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("Hello world")), nil),
			func(_ *http.Request) (string, int, map[string]interface{}, error) {
				return "error", http.StatusUnauthorized, nil, nil
			},
			`messageContent=Hello+world&messageLevel=success`,
			http.StatusUnauthorized,
			http.Header{},
		},
		{
			"error",
			configuredApp,
			httptest.NewRequest(http.MethodGet, "/", nil),
			func(_ *http.Request) (string, int, map[string]interface{}, error) {
				return "", http.StatusBadRequest, nil, model.WrapInvalid(errors.New("error"))
			},
			`messageContent=error%3A+invalid&messageLevel=error`,
			http.StatusBadRequest,
			http.Header{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.Handler(tc.templateFunc).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	var cases = []struct {
		intention  string
		instance   app
		request    *http.Request
		path       string
		message    Message
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			app{
				publicURL: "http://vibioh.fr",
			},
			httptest.NewRequest(http.MethodGet, "http://vibioh.fr/", nil),
			"/",
			NewSuccessMessage("Created with success"),
			"<a href=\"http://vibioh.fr/?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("http://vibioh.fr/?%s", NewSuccessMessage("Created with success"))},
			},
		},
		{
			"relative URL",
			app{
				publicURL: "http://localhost:1080",
			},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success",
			NewSuccessMessage("Created with success"),
			"<a href=\"http://localhost:1080/success?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("http://localhost:1080/success?%s", NewSuccessMessage("Created with success"))},
			},
		},
		{
			"exact URL",
			app{
				publicURL: "https://vibioh.fr",
			},
			httptest.NewRequest(http.MethodGet, "https://vibioh.fr", nil),
			"https://app.local/success",
			NewSuccessMessage("Created with success"),
			"<a href=\"https://app.local/success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{"https://app.local/success"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.Redirect(writer, tc.request, tc.path, tc.message)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Redirect = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Redirect = `%s`, want `%s`", string(got), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}

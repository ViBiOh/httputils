package renderer

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/renderer/model"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -publicURL string\n    \tPublic URL {SIMPLE_PUBLIC_URL} (default \"http://localhost\")\n  -static string\n    \tStatic folder, content served directly {SIMPLE_STATIC} (default \"./static/\")\n  -templates string\n    \tHTML Templates folder {SIMPLE_TEMPLATES} (default \"./templates/\")\n  -title string\n    \tApplication title {SIMPLE_TITLE} (default \"App\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
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

func TestHandler(t *testing.T) {
	templates := "../../templates/"
	statics := "../../templates/static/"
	publicURL := "http://localhost"
	title := "Golang Test"

	configuredApp, err := New(Config{
		templates: &templates,
		statics:   &statics,
		publicURL: &publicURL,
		title:     &title,
	}, template.FuncMap{})
	if err != nil {
		t.Error(err)
	}

	var cases = []struct {
		intention    string
		instance     App
		request      *http.Request
		templateFunc model.TemplateFunc
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
			httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", model.NewSuccessMessage("Hello world")), nil),
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
		request    *http.Request
		path       string
		message    model.Message
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"/success",
			model.NewSuccessMessage("Created with success"),
			"<a href=\"/success?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?%s", model.NewSuccessMessage("Created with success"))},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Redirect(writer, tc.request, tc.path, tc.message)

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

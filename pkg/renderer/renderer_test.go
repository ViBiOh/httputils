package renderer

import (
	"context"
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
	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -minify\n    \tMinify HTML ${SIMPLE_MINIFY} (default true)\n  -pathPrefix string\n    \tRoot Path Prefix ${SIMPLE_PATH_PREFIX}\n  -publicURL string\n    \tPublic URL ${SIMPLE_PUBLIC_URL} (default \"http://localhost:1080\")\n  -title string\n    \tApplication title ${SIMPLE_TITLE} (default \"App\")\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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

func TestFeedContent(t *testing.T) {
	t.Parallel()

	type args struct {
		content map[string]any
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     map[string]any
	}{
		"empty": {
			Service{},
			args{
				content: nil,
			},
			make(map[string]any),
		},
		"merge": {
			Service{
				content: map[string]any{
					"Version": "test",
				},
			},
			args{
				content: map[string]any{
					"Name": "Hello World",
				},
			},
			map[string]any{
				"Version": "test",
				"Name":    "Hello World",
			},
		},
		"no overwrite": {
			Service{
				content: map[string]any{
					"Title": "test",
				},
			},
			args{
				content: map[string]any{
					"Title": "Hello World",
				},
			},
			map[string]any{
				"Title": "Hello World",
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.feedContent(testCase.args.content); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("feedContent() = %v, want %v", testCase.args.content, testCase.want)
			}
		})
	}
}

//go:embed templates static
var content embed.FS

func TestHandler(t *testing.T) {
	t.Parallel()

	configuredService, err := New(context.Background(), &Config{
		PublicURL: "http://localhost",
		Title:     "Golang Test",
		Minify:    true,
	}, content, template.FuncMap{}, nil, nil)
	if err != nil {
		t.Error(err)
	}

	cases := map[string]struct {
		instance     *Service
		request      *http.Request
		templateFunc TemplateFunc
		want         string
		wantStatus   int
		wantHeader   http.Header
	}{
		"html": {
			configuredService,
			httptest.NewRequest(http.MethodGet, "/", nil),
			func(_ http.ResponseWriter, _ *http.Request) (Page, error) {
				return Page{
					Template: "public",
					Status:   http.StatusCreated,
					Content:  nil,
				}, nil
			},
			`<!doctype html><html lang=en><meta charset=utf-8><title>Golang Test</title>
<h1>Hello !</h1>`,
			http.StatusCreated,
			http.Header{},
		},
		"message": {
			configuredService,
			httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("Hello world")), nil),
			func(_ http.ResponseWriter, _ *http.Request) (Page, error) {
				return Page{
					Template: "error",
					Status:   http.StatusUnauthorized,
					Content:  nil,
				}, nil
			},
			`msgKey=Message&amp;msgCnt=Hello+world&amp;msgLvl=success`,
			http.StatusUnauthorized,
			http.Header{},
		},
		"error": {
			configuredService,
			httptest.NewRequest(http.MethodGet, "/", nil),
			func(_ http.ResponseWriter, _ *http.Request) (Page, error) {
				return Page{
					Template: "",
					Status:   http.StatusBadRequest,
					Content:  nil,
				}, model.WrapInvalid(errors.New("error"))
			},
			`msgKey=Message&amp;msgCnt=error%0Ainvalid&amp;msgLvl=error`,
			http.StatusBadRequest,
			http.Header{},
		},
		"unknown": {
			configuredService,
			httptest.NewRequest(http.MethodGet, "/", nil),
			func(_ http.ResponseWriter, _ *http.Request) (Page, error) {
				return Page{
					Template: "unknownpage",
					Status:   http.StatusOK,
					Content:  nil,
				}, nil
			},
			`msgKey=Message&amp;msgCnt=unknown+template+%60unknownpage%60%0Anot+found&amp;msgLvl=error`,
			http.StatusNotFound,
			http.Header{},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.Handler(testCase.templateFunc).ServeHTTP(writer, testCase.request)

			assert.Equal(t, testCase.wantStatus, writer.Code)

			actual, _ := request.ReadBodyResponse(writer.Result())
			assert.Equal(t, testCase.want, string(actual))

			for key := range testCase.wantHeader {
				assert.Equal(t, testCase.wantHeader.Get(key), writer.Header().Get(key))
			}
		})
	}
}

func TestHandleStatic(t *testing.T) {
	t.Parallel()

	configuredService, err := New(context.Background(), &Config{
		PublicURL: "http://localhost",
		Title:     "Golang Test",
		Minify:    true,
	}, content, template.FuncMap{}, nil, nil)
	if err != nil {
		t.Error(err)
	}

	faviconRequest := httptest.NewRequest(http.MethodGet, "/images/favicon/manifest.json", nil)
	faviconRequest.SetPathValue("path", "images/favicon/manifest.json")

	cases := map[string]struct {
		instance   *Service
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		"favicon": {
			configuredService,
			faviconRequest,
			"{}\n",
			http.StatusOK,
			http.Header{},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.HandleStatic("/").ServeHTTP(writer, testCase.request)

			assert.Equal(t, testCase.wantStatus, writer.Code)

			actual, _ := request.ReadBodyResponse(writer.Result())
			assert.Equal(t, testCase.want, string(actual))

			for key := range testCase.wantHeader {
				assert.Equal(t, testCase.wantHeader.Get(key), writer.Header().Get(key))
			}
		})
	}
}

func TestHandleSVG(t *testing.T) {
	t.Parallel()

	configuredService, err := New(context.Background(), &Config{
		PublicURL: "http://localhost",
		Title:     "Golang Test",
		Minify:    true,
	}, content, template.FuncMap{}, nil, nil)
	if err != nil {
		t.Error(err)
	}

	configuredPrefixService, err := New(context.Background(), &Config{
		PublicURL:  "http://localhost",
		PathPrefix: "/app",
		Title:      "Golang Test",
		Minify:     true,
	}, content, template.FuncMap{}, nil, nil)
	if err != nil {
		t.Error(err)
	}

	svgTestRequest := httptest.NewRequest(http.MethodGet, "/svg/test?fill=black", nil)
	svgTestRequest.SetPathValue("path", "test")

	svgAppTestRequest := httptest.NewRequest(http.MethodGet, "/app/svg/test?fill=black", nil)
	svgAppTestRequest.SetPathValue("path", "test")

	unknownRequest := httptest.NewRequest(http.MethodGet, "/svg/unknown?fill=black", nil)
	unknownRequest.SetPathValue("path", "unknown")

	cases := map[string]struct {
		instance   *Service
		request    *http.Request
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		"svg": {
			configuredService,
			svgTestRequest,
			"color=black",
			http.StatusOK,
			http.Header{},
		},
		"svg with prefix": {
			configuredPrefixService,
			svgAppTestRequest,
			"color=black",
			http.StatusOK,
			http.Header{},
		},
		"svg not found": {
			configuredService,
			unknownRequest,
			"🤷\n",
			http.StatusNotFound,
			http.Header{},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.HandleSVG().ServeHTTP(writer, testCase.request)

			assert.Equal(t, testCase.wantStatus, writer.Code)

			actual, _ := request.ReadBodyResponse(writer.Result())
			assert.Equal(t, testCase.want, string(actual))

			for key := range testCase.wantHeader {
				assert.Equal(t, testCase.wantHeader.Get(key), writer.Header().Get(key))
			}
		})
	}
}

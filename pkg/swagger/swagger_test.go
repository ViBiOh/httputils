package swagger

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
)

var (
	errSwagger = errors.New("unable to generate config")
)

func swaggerError() (Configuration, error) {
	return EmptyConfiguration, errSwagger
}

func swaggerEmpty() (Configuration, error) {
	return EmptyConfiguration, nil
}

func swaggerBasic() (Configuration, error) {
	return Configuration{
		Paths: `/health:
  get:
    description: Healthcheck of app
    responses:
      200:
        description: Everything is fine`,
		Components: `Error:
  description: Plain text Error
  content:
    text/plain:
      schema:
        type: string`,
	}, nil
}

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -title string\n    \t[swagger] API Title {SIMPLE_TITLE} (default \"API\")\n  -version string\n    \t[swagger] API Version {SIMPLE_VERSION} (default \"1.0.0\")\n",
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

func TestNew(t *testing.T) {
	input := "hello"

	var cases = []struct {
		intention string
		config    Config
		providers []Provider
		want      App
		wantErr   error
	}{
		{
			"simple",
			Config{
				title:   &input,
				version: &input,
			},
			nil,
			&app{
				swaggerContent: []byte(`---

openapi: 3.0.0
info:
  description: API for hello
  title: hello
  version: hello

paths:

components:
  schemas:
`),
				data: map[string]string{
					"Title": "hello",
				},
			},
			nil,
		},
		{
			"swagger error",
			Config{
				title:   &input,
				version: &input,
			},
			[]Provider{swaggerError},
			nil,
			errSwagger,
		},
		{
			"configuration management",
			Config{
				title:   &input,
				version: &input,
			},
			[]Provider{swaggerEmpty, swaggerBasic},

			&app{
				swaggerContent: []byte(`---

openapi: 3.0.0
info:
  description: API for hello
  title: hello
  version: hello

paths:
  /health:
    get:
      description: Healthcheck of app
      responses:
        200:
          description: Everything is fine
components:
  schemas:
    Error:
      description: Plain text Error
      content:
        text/plain:
          schema:
            type: string`),
				data: map[string]string{
					"Title": "hello",
				},
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := New(testCase.config, testCase.providers...)

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("New() = (%v, `%s`), want (%v, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	input := "hello"
	instance, _ := New(Config{
		title:   &input,
		version: &input,
	})

	var cases = []struct {
		intention  string
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"index",
			instance,
			httptest.NewRequest(http.MethodGet, "/", nil),
			`<!doctype html><html class=no-js lang=en><meta charset=utf-8><meta http-equiv=x-ua-compatible content="IE=edge,chrome=1"><title>Swagger hello</title><meta name=description content="Swagger UI of hello"><meta name=viewport content="width=device-width,initial-scale=1"><link rel=stylesheet href=//unpkg.com/swagger-ui-dist@3/swagger-ui.css><div id=swagger-ui></div><script src=//unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js></script><script>
      SwaggerUIBundle({
        url: "swagger.yaml",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis]
      })
    </script>`,
			http.StatusOK,
		},
		{
			"swagger",
			instance,
			httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil),
			`---

openapi: 3.0.0
info:
  description: API for hello
  title: hello
  version: hello

paths:

components:
  schemas:
`,
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			testCase.instance.Handler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Handler = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Handler = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

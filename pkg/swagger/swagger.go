package swagger

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

const (
	indexTemplateStr = `<!doctype html>
<html class="no-js" lang="">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <title>Swagger {{ .Title }}</title>
    <meta name="description" content="Swagger UI of {{ .Title }}">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" type="text/css" href="//unpkg.com/swagger-ui-dist@3/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="//unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
    <script>
      SwaggerUIBundle({
        url: "swagger.yaml",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis]
      })
    </script>
  </body>
</html>`

	swaggerTemplateStr = `---

openapi: 3.0.0
info:
  description: API for {{ .Title }}
  title: {{ .Title }}
  version: '1.0.0'

paths:
  /health:
    get:
      description: Healthcheck of app
      responses:
        '204':
          description: Everything is fine

  /version:
    get:
      description: Version of app

      responses:
        '200':
          description: Version of app
          content:
            text/plain:
              schema:
                type: string
{{ .Paths }}
components:
  schemas:
{{ .Components }}
    Error:
      description: Request Error
      content:
        text/plain:
          schema:
            type: string
`
)

var (
	indextTemplate  *template.Template
	swaggerTemplate *template.Template
)

func init() {
	tpl, err := template.New("index").Parse(indexTemplateStr)
	logger.Fatal(err)
	indextTemplate = tpl

	tpl, err = template.New("swagger").Parse(swaggerTemplateStr)
	logger.Fatal(err)
	swaggerTemplate = tpl
}

// App of package
type App interface {
	Handler() http.Handler
}

// Config of package
type Config struct {
	title *string
}

type app struct {
	data           map[string]string
	swaggerContent []byte
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		title: flags.New(prefix, "swagger").Name("Title").Default("API").Label("API Title").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, cruds ...crud.App) (App, error) {
	paths := strings.Builder{}
	components := strings.Builder{}

	for _, crudApp := range cruds {
		path, component, err := crudApp.Swagger()
		if err == crud.ErrNoSwaggerConfiguration {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("unable to generate swagger for %v: %s", crudApp, err)
		}

		paths.WriteString(path)
		paths.WriteString("\n")
		components.WriteString(component)
		components.WriteString("\n")
	}

	data := map[string]string{
		"Title":      strings.TrimSpace(*config.title),
		"Paths":      paths.String(),
		"Components": components.String(),
	}

	swaggerContent := bytes.Buffer{}
	if err := swaggerTemplate.Execute(&swaggerContent, data); err != nil {
		return nil, err
	}

	return &app{
		swaggerContent: swaggerContent.Bytes(),
		data:           data,
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/swagger.yaml") {
			w.Header().Set("Content-Type", "application/x-yaml; charset=UTF-8")
			w.WriteHeader(http.StatusOK)

			if _, err := w.Write(a.swaggerContent); err != nil {
				logger.Error("unable to write swagger: %s", err)
			}
			return
		}

		if err := templates.WriteHTMLTemplate(indextTemplate, w, a.data, http.StatusOK); err != nil {
			logger.Error("unable to write index: %s", err)
		}
	})
}

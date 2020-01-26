package swagger

import (
	"bytes"
	"flag"
	"fmt"
	htmlTemplate "html/template"
	"net/http"
	"regexp"
	"strings"
	textTemplate "text/template"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/query"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

// Configuration holds configuration parts
type Configuration struct {
	Paths      string
	Components string
}

// Provider provides configuration
type Provider func() (Configuration, error)

const (
	indexTemplateStr = `<!doctype html>
<html class="no-js" lang="en">
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
  version: {{ .Version }}

paths:
{{ .Paths }}
components:
  schemas:
{{ .Components }}`
)

var (
	// EmptyConfiguration is an empty configuration
	EmptyConfiguration = Configuration{}

	indextTemplate  *htmlTemplate.Template
	swaggerTemplate *textTemplate.Template
	prefixer        = regexp.MustCompile(`(?m)^(.+)$`)
)

func init() {
	index, err := htmlTemplate.New("index").Parse(indexTemplateStr)
	logger.Fatal(err)
	indextTemplate = index

	swagger, err := textTemplate.New("swagger").Parse(swaggerTemplateStr)
	logger.Fatal(err)
	swaggerTemplate = swagger
}

// App of package
type App interface {
	Handler() http.Handler
}

// Config of package
type Config struct {
	title   *string
	version *string
}

type app struct {
	data           map[string]string
	swaggerContent []byte
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		title:   flags.New(prefix, "swagger").Name("Title").Default("API").Label("API Title").ToString(fs),
		version: flags.New(prefix, "swagger").Name("Version").Default("1.0.0").Label("API Version").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, providers ...Provider) (App, error) {
	paths := make([]string, 0)
	components := make([]string, 0)

	for _, provider := range providers {
		configuration, err := provider()
		if err != nil {
			return nil, fmt.Errorf("unable to generate swagger for %T: %w", provider, err)
		}

		if configuration == EmptyConfiguration {
			continue
		}

		paths = append(paths, prefixer.ReplaceAllString(configuration.Paths, "  $1"))
		components = append(components, prefixer.ReplaceAllString(configuration.Components, "    $1"))
	}

	data := map[string]string{
		"Title":      strings.TrimSpace(*config.title),
		"Version":    strings.TrimSpace(*config.version),
		"Paths":      strings.Join(paths, "\n"),
		"Components": strings.Join(components, "\n"),
	}

	swaggerContent := bytes.Buffer{}
	if err := swaggerTemplate.Execute(&swaggerContent, data); err != nil {
		return nil, fmt.Errorf("unable to execute swagger template: %s", err)
	}

	return &app{
		swaggerContent: swaggerContent.Bytes(),
		data: map[string]string{
			"Title": strings.TrimSpace(*config.title),
		},
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if query.IsRoot(r) {
			if err := templates.ResponseHTMLTemplate(indextTemplate, w, a.data, http.StatusOK); err != nil {
				logger.Error("unable to write index: %s", err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(a.swaggerContent); err != nil {
			logger.Error("unable to write swagger: %s", err)
		}
		return
	})
}

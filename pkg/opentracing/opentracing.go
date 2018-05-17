package opentracing

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"
	opentracing "github.com/opentracing/opentracing-go"
)

// App stores informations
type App struct {
	name string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	serviceName := strings.TrimSpace(*config[`service`])
	if serviceName == `` {
		log.Print(`[opentracing] ⚠ No service name provided`)
		return &App{}
	}

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	return &App{
		name: serviceName,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`name`: flag.String(tools.ToCamel(fmt.Sprintf(`%sName`, prefix)), ``, `[opentracing] Service name`),
	}
}

func (a App) check() bool {
	return a.name != ``
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	if !a.check() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

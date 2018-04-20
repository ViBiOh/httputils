package datadog

import (
	"flag"
	"fmt"
	"net/http"

	ddhttp "github.com/DataDog/dd-trace-go/contrib/net/http"
	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/DataDog/dd-trace-go/tracer/ext"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// App stores informations
type App struct {
	serviceName string
	tracer      *tracer.Tracer
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	serviceName := *config[`service`]

	if serviceName == `` {
		return &App{}
	}

	ddTracer := tracer.NewTracerTransport(tracer.NewTransport(*config[`hostname`], *config[`port`]))
	ddTracer.SetServiceInfo(serviceName, `net/http`, ext.AppTypeWeb)

	return &App{
		serviceName: serviceName,
		tracer:      ddTracer,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`service`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sService`, prefix)), ``, `Service name`),
		`hostname`: flag.String(tools.ToCamel(fmt.Sprintf(`%sHostname`, prefix)), `dd-agent`, `Datadog Agent Hostname`),
		`port`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), `8126`, `Datadog Agent Port`),
	}
}

// Handler for net/http
func (a *App) Handler(next http.Handler) http.Handler {
	if a.tracer == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := a.tracer.NewRootSpan(`http.request`, a.serviceName, fmt.Sprintf(`%s %s`, r.Method, r.URL.Path))
		defer span.Finish()

		// Configuration span
		span.Type = ext.HTTPType
		span.SetMeta(ext.HTTPMethod, r.Method)
		span.SetMeta(ext.HTTPURL, r.URL.Path)

		// Enriching request and writer with tracer
		ctx := span.Context(r.Context())
		traceRequest := r.WithContext(ctx)
		traceWriter := ddhttp.NewResponseWriter(w, span)

		next.ServeHTTP(traceWriter, traceRequest)
	})
}

package opentracing

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/tools"
	opentracing "github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// App stores informations
type App struct {
	tracer opentracing.Tracer
	closer io.Closer
}

func initJaeger(serviceName string, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
	config := jaegercfg.Configuration{
		ServiceName: serviceName,
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := config.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)

	if err != nil {
		return nil, nil, fmt.Errorf(`Error while initializing Jaeger tracer: %v`, err)
	}

	return tracer, closer, nil
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	serviceName := strings.TrimSpace(*config[`service`])
	if serviceName == `` {
		log.Print(`[opentracing] ⚠ No service name provided`)
		return &App{}
	}

	tracer, closer, err := initJaeger(serviceName, strings.TrimSpace(*config[`agent`]))
	if err != nil {
		log.Printf(`[opentracing] %v`, err)
		if closer != nil {
			defer func() {
				if err := closer.Close(); err != nil {
					log.Printf(`[opentracing] Error while closing tracer: %v`, err)
				}
			}()
		}

		return &App{}
	}

	opentracing.SetGlobalTracer(tracer)

	return &App{
		tracer: tracer,
		closer: closer,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`name`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sName`, prefix)), ``, `[opentracing] Service name`),
		`agent`: flag.String(tools.ToCamel(fmt.Sprintf(`%sAgent`, prefix)), `jaeger:6831`, `[opentracing] Jaeger Agent host:port`),
	}
}

func (a App) check() bool {
	return a.tracer != nil
}

// Close tracer
func (a App) Close() {
	if !a.check() {
		return
	}

	if err := a.closer.Close(); err != nil {
		log.Printf(`[opentracing] Error while closing tracer: %v`, err)
	}
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	if !a.check() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, _ := opentracing.StartSpanFromContext(r.Context(), `http.request`)
		defer span.Finish()

		span.LogFields(
			opentracingLog.String(`http.method`, r.Method),
			opentracingLog.String(`http.url`, r.URL.Path),
			opentracingLog.String(`http.remote_addr`, r.RemoteAddr),
			opentracingLog.String(`headers.real_ip`, r.Header.Get(`X-Real-Ip`)),
			opentracingLog.String(`headers.forwarded_for`, r.Header.Get(`X-Forwarded-For`)),
			opentracingLog.String(`headers.user_agent`, r.Header.Get(`User-Agent`)),
		)

		next.ServeHTTP(w, r)
	})
}

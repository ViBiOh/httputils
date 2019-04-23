package opentracing

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

var _ model.Middleware = &App{}

// Config of package
type Config struct {
	name  *string
	agent *string
}

// App of package
type App struct {
	closer io.Closer
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "opentracing"
	}

	return Config{
		name:  fs.String(tools.ToCamel(fmt.Sprintf("%sName", prefix)), "", fmt.Sprintf("[%s] Service name", docPrefix)),
		agent: fs.String(tools.ToCamel(fmt.Sprintf("%sAgent", prefix)), "jaeger:6831", fmt.Sprintf("[%s] Jaeger Agent (e.g. host:port)", docPrefix)),
	}
}

// New creates new App from Config
func New(config Config) *App {
	serviceName := strings.TrimSpace(*config.name)
	if serviceName == "" {
		logger.Warn("no service name provided")
		return &App{}
	}

	tracer, closer, err := initJaeger(serviceName, strings.TrimSpace(*config.agent))
	if err != nil {
		logger.Error("%+v", err)
		if closer != nil {
			defer func() {
				if err := closer.Close(); err != nil {
					logger.Error("%+v", errors.WithStack(err))
				}
			}()
		}

		return &App{}
	}

	opentracing.SetGlobalTracer(tracer)

	return &App{
		closer: closer,
	}
}

func initJaeger(serviceName string, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
	config := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}

	tracer, closer, err := config.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return tracer, closer, nil
}

func (a App) check() bool {
	return a.closer != nil
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	if !a.check() {
		return next
	}

	return nethttp.Middleware(opentracing.GlobalTracer(), next, nethttp.MWSpanObserver(func(span opentracing.Span, r *http.Request) {
		span.SetTag("http.remote_addr", r.RemoteAddr)
		span.SetTag("headers.real_ip", r.Header.Get("X-Real-Ip"))
		span.SetTag("headers.forwarded_for", r.Header.Get("X-Forwarded-For"))
		span.SetTag("headers.user_agent", r.Header.Get("User-Agent"))
	}))
}

// Close tracer
func (a App) Close() {
	if a.closer == nil {
		return
	}

	if err := a.closer.Close(); err != nil {
		logger.Error("%+v", errors.WithStack(err))
	}
}

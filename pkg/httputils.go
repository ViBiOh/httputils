package httputils

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// Config of package
type Config struct {
	port             *int
	gracefulDuration *string
	cert             *string
	key              *string
}

// App of package
type App struct {
	port             int
	gracefulDuration time.Duration
	cert             string
	key              string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "http"
	}

	return Config{
		port:             fs.Int(tools.ToCamel(fmt.Sprintf("%sPort", prefix)), 1080, fmt.Sprintf("[%s] Listen port", docPrefix)),
		gracefulDuration: fs.String(tools.ToCamel(fmt.Sprintf("%sGraceful", prefix)), "35s", fmt.Sprintf("[%s] Graceful close duration", docPrefix)),
		cert:             fs.String(tools.ToCamel(fmt.Sprintf("%sCert", prefix)), "", fmt.Sprintf("[%s] Certificate file", docPrefix)),
		key:              fs.String(tools.ToCamel(fmt.Sprintf("%sKey", prefix)), "", fmt.Sprintf("[%s] Key file", docPrefix)),
	}
}

// New creates new App from Config
func New(config Config) (*App, error) {
	gracefulDuration, err := time.ParseDuration(*config.gracefulDuration)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &App{
		port:             *config.port,
		gracefulDuration: gracefulDuration,
		cert:             *config.cert,
		key:              *config.key,
	}, nil
}

// VersionHandler for sending current app version from `VERSION` environment variable
func VersionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if _, err := w.Write([]byte(os.Getenv("VERSION"))); err != nil {
			logger.Error("%#v", errors.WithStack(err))
		}
	})
}

// ListenAndServe starts server
func (a App) ListenAndServe(handler http.Handler, onShutdown func(), healthcheckApp *healthcheck.App, flushers ...model.Flusher) {
	healthcheckHandler := healthcheckApp.Handler()
	versionHandler := VersionHandler()

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				healthcheckHandler.ServeHTTP(w, r)
			case "/version":
				versionHandler.ServeHTTP(w, r)
			default:
				handler.ServeHTTP(w, r)
			}
		}),
	}

	if onShutdown != nil {
		httpServer.RegisterOnShutdown(onShutdown)
	}

	logger.Info("Starting HTTP server on port %s", httpServer.Addr)

	var serveError = make(chan error)
	go func() {
		defer close(serveError)

		if a.cert != "" && a.key != "" {
			logger.Info("Listening with TLS")
			serveError <- httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			logger.Warn("Listening without TLS")
			serveError <- httpServer.ListenAndServe()
		}
	}()

	server.GracefulClose(httpServer, a.gracefulDuration, serveError, healthcheckApp, flushers...)
}

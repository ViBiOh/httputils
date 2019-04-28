package httputils

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// Config of package
type Config struct {
	port *int
	cert *string
	key  *string
}

// App of package
type App struct {
	port int
	cert string
	key  string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "http"
	}

	return Config{
		port: fs.Int(tools.ToCamel(fmt.Sprintf("%sPort", prefix)), 1080, fmt.Sprintf("[%s] Listen port", docPrefix)),
		cert: fs.String(tools.ToCamel(fmt.Sprintf("%sCert", prefix)), "", fmt.Sprintf("[%s] Certificate file", docPrefix)),
		key:  fs.String(tools.ToCamel(fmt.Sprintf("%sKey", prefix)), "", fmt.Sprintf("[%s] Key file", docPrefix)),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		port: *config.port,
		cert: *config.cert,
		key:  *config.key,
	}
}

// ListenAndServe starts server
func (a App) ListenAndServe(handler http.Handler, onGracefulClose func() error, healthcheckApp *healthcheck.App, flushers ...model.Flusher) {
	healthcheckHandler := healthcheckApp.Handler()

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				healthcheckHandler.ServeHTTP(w, r)
			} else {
				handler.ServeHTTP(w, r)
			}
		}),
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

	server.GracefulClose(httpServer, serveError, onGracefulClose, healthcheckApp, flushers...)
}

package httputils

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/cert"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// Config of package
type Config struct {
	port       *int
	tls        *bool
	certConfig cert.Config
}

// App of package
type App struct {
	port       int
	tls        bool
	certConfig cert.Config
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		port:       fs.Int(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), 1080, `Listen port`),
		tls:        flag.Bool(tools.ToCamel(fmt.Sprintf(`%sTls`, prefix)), true, `Serve TLS content`),
		certConfig: cert.Flags(fs, tools.ToCamel(fmt.Sprintf(`%sTls`, prefix))),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		port:       *config.port,
		tls:        *config.tls,
		certConfig: config.certConfig,
	}
}

// ListenAndServe starts server
func (a App) ListenAndServe(handler http.Handler, onGracefulClose func() error, healthcheckApp *healthcheck.App, flushers ...model.Flusher) {
	healthcheckHandler := healthcheckApp.Handler()

	httpServer := &http.Server{
		Addr: fmt.Sprintf(`:%d`, a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == `/health` {
				healthcheckHandler.ServeHTTP(w, r)
			} else {
				handler.ServeHTTP(w, r)
			}
		}),
	}

	logger.Info(`Starting HTTP server on port %s`, httpServer.Addr)

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if a.tls {
			logger.Info(`Listening with TLS`)
			serveError <- cert.ListenAndServeTLS(a.certConfig, httpServer)
		} else {
			logger.Warn(`Listening without TLS`)
			serveError <- httpServer.ListenAndServe()
		}
	}()

	server.GracefulClose(httpServer, serveError, onGracefulClose, healthcheckApp, flushers...)
}

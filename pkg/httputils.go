package httputils

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	httpShutdownTimeout = 10 * time.Second
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

		version := os.Getenv("VERSION")
		if version == "" {
			version = "development"
		}

		if _, err := w.Write([]byte(version)); err != nil {
			logger.Error("%#v", errors.WithStack(err))
		}
	})
}

// HealthHandler for dealing with state of app
func HealthHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if handler != nil {
				handler.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// ChainMiddlewares chains middlewares call for easy wrapping
func ChainMiddlewares(handler http.Handler, middlewares ...model.Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i].Handler(result)
	}

	return result
}

// ListenAndServe starts server
func (a App) ListenAndServe(handler http.Handler, healthHandler http.Handler, onShutdown func()) {
	versionHandler := VersionHandler()

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				healthHandler.ServeHTTP(w, r)
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

	var err error

	if a.cert != "" && a.key != "" {
		logger.Info("Listening with TLS")
		err = httpServer.ListenAndServeTLS(a.cert, a.key)
	} else {
		logger.Warn("Listening without TLS")
		err = httpServer.ListenAndServe()
	}

	if err != nil {
		logger.Error("%#v", errors.WithStack(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		logger.Error("%#v", errors.WithStack(err))
	}
}

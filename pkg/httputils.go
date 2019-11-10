package httputils

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/model"
)

// App of package
type App interface {
	ListenAndServe(http.Handler, http.Handler) (*http.Server, <-chan error)
	ListenServeWait(http.Handler, http.Handler)
}

// Config of package
type Config struct {
	address *string
	port    *int
	cert    *string
	key     *string
}

type app struct {
	address          string
	port             int
	gracefulDuration time.Duration
	cert             string
	key              string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		address: flags.New(prefix, "http").Name("Address").Default("").Label("Listen address").ToString(fs),
		port:    flags.New(prefix, "http").Name("Port").Default(1080).Label("Listen port").ToInt(fs),
		cert:    flags.New(prefix, "http").Name("Cert").Default("").Label("Certificate file").ToString(fs),
		key:     flags.New(prefix, "http").Name("Key").Default("").Label("Key file").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return &app{
		address: *config.address,
		port:    *config.port,
		cert:    *config.cert,
		key:     *config.key,
	}
}

func versionHandler() http.Handler {
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
			logger.Error("%s", err)
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
			return

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
func (a app) ListenAndServe(handler http.Handler, healthHandler http.Handler) (*http.Server, <-chan error) {
	versionHandler := versionHandler()
	if healthHandler == nil {
		healthHandler = HealthHandler(nil)
	}

	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", a.address, a.port),
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

	logger.Info("Starting HTTP server on %s", httpServer.Addr)

	err := make(chan error, 1)

	go func() {
		if a.cert != "" && a.key != "" {
			logger.Info("Listening with TLS")
			err <- httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			logger.Warn("Listening without TLS")
			err <- httpServer.ListenAndServe()
		}
	}()

	return httpServer, err
}

// ListenServeWait starts server and wait for its termination
func (a app) ListenServeWait(handler http.Handler, healthHandler http.Handler) {
	_, err := a.ListenAndServe(handler, healthHandler)
	WaitForTermination(err)
}

// WaitForTermination wait for error or SIGTERM/SIGINT signal
func WaitForTermination(err <-chan error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-err:
		logger.Error("%s", err)
	case signal := <-signals:
		logger.Info("%s received", signal)
	}
}

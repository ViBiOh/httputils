package httputils

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/model"
)

// App of package
type App interface {
	Health(http.Handler) App
	Middleware(model.Middleware) App
	ListenAndServe(http.Handler) (*http.Server, <-chan error)
	ListenServeWait(http.Handler)
}

// Config of package
type Config struct {
	address  *string
	port     *int
	cert     *string
	key      *string
	okStatus *int
}

type app struct {
	address  string
	port     int
	cert     string
	key      string
	okStatus int

	middlewares []model.Middleware
	health      http.Handler
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		address:  flags.New(prefix, "http").Name("Address").Default("").Label("Listen address").ToString(fs),
		port:     flags.New(prefix, "http").Name("Port").Default(1080).Label("Listen port").ToInt(fs),
		cert:     flags.New(prefix, "http").Name("Cert").Default("").Label("Certificate file").ToString(fs),
		key:      flags.New(prefix, "http").Name("Key").Default("").Label("Key file").ToString(fs),
		okStatus: flags.New(prefix, "http").Name("OkStatus").Default(http.StatusNoContent).Label("Healthy HTTP Status code").ToInt(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return &app{
		address: *config.address,
		port:    *config.port,
		cert:    *config.cert,
		key:     *config.key,

		health:      HealthHandler(*config.okStatus),
		middlewares: make([]model.Middleware, 0),
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
func HealthHandler(okStatus int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(okStatus)
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

// Health set health http handler
func (a *app) Health(health http.Handler) App {
	a.health = health

	return a
}

// Middleware add given middleware to list
func (a *app) Middleware(middleware model.Middleware) App {
	a.middlewares = append(a.middlewares, middleware)

	return a
}

// ListenAndServe starts server
func (a *app) ListenAndServe(handler http.Handler) (*http.Server, <-chan error) {
	versionHandler := versionHandler()
	defaultHandler := ChainMiddlewares(handler, a.middlewares...)

	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", a.address, a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				a.health.ServeHTTP(w, r)
			case "/version":
				versionHandler.ServeHTTP(w, r)
			default:
				defaultHandler.ServeHTTP(w, r)
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
func (a *app) ListenServeWait(handler http.Handler) {
	_, err := a.ListenAndServe(handler)
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

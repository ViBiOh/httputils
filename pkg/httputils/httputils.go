package httputils

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
)

var (
	_ swagger.Provider = (&app{}).Swagger
)

// Pinger describes a function to check liveness of app
type Pinger = func() error

// App of package
type App interface {
	Health(Pinger) App
	Middleware(model.Middleware) App
	ListenAndServe(http.Handler) (*http.Server, <-chan error)
	ListenServeWait(http.Handler)
	Swagger() (swagger.Configuration, error)
}

// Config of package
type Config struct {
	address       *string
	port          *uint
	cert          *string
	key           *string
	graceDuration *string
	okStatus      *int
}

type app struct {
	listenAddress string
	cert          string
	key           string
	okStatus      int
	graceDuration time.Duration

	shutdown    bool
	middlewares []model.Middleware
	pingers     []Pinger
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		address:       flags.New(prefix, "http").Name("Address").Default("").Label("Listen address").ToString(fs),
		port:          flags.New(prefix, "http").Name("Port").Default(1080).Label("Listen port").ToUint(fs),
		cert:          flags.New(prefix, "http").Name("Cert").Default("").Label("Certificate file").ToString(fs),
		key:           flags.New(prefix, "http").Name("Key").Default("").Label("Key file").ToString(fs),
		graceDuration: flags.New(prefix, "http").Name("GraceDuration").Default("15s").Label("Grace duration when SIGTERM received").ToString(fs),
		okStatus:      flags.New(prefix, "http").Name("OkStatus").Default(http.StatusNoContent).Label("Healthy HTTP Status code").ToInt(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	graceDurationValue := strings.TrimSpace(*config.graceDuration)
	graceDuration, err := time.ParseDuration(graceDurationValue)
	if err != nil {
		logger.Warn("invalid graceDuration `%s`: %s", graceDurationValue, err)
	}

	return &app{
		listenAddress: fmt.Sprintf("%s:%d", *config.address, *config.port),
		cert:          *config.cert,
		key:           *config.key,
		okStatus:      *config.okStatus,
		graceDuration: graceDuration,

		middlewares: make([]model.Middleware, 0),
		pingers:     make([]Pinger, 0),
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

// Health set health http handler
func (a *app) healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if a.shutdown {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		for _, pinger := range a.pingers {
			if err := pinger(); err != nil {
				logger.Error("unable to ping: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		w.WriteHeader(a.okStatus)
	})
}

// ChainMiddlewares chains middlewares call for easy wrapping
func ChainMiddlewares(handler http.Handler, middlewares ...model.Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}

	return result
}

// Middleware add given middleware to list
func (a *app) Middleware(middleware model.Middleware) App {
	a.middlewares = append(a.middlewares, middleware)

	return a
}

// Health add given pinger to list
func (a *app) Health(pinger Pinger) App {
	a.pingers = append(a.pingers, pinger)

	return a
}

func (a *app) Swagger() (swagger.Configuration, error) {
	paths := fmt.Sprintf(`/health:
  get:
    description: Healthcheck of app
    responses:
      %d:
        description: Everything is fine

/version:
  get:
    description: Version of app

    responses:
      200:
        description: Version of app
        content:
          text/plain:
            schema:
              type: string`, a.okStatus)

	return swagger.Configuration{
		Paths: paths,
		Components: `Error:
  description: Plain text Error
  content:
    text/plain:
      schema:
        type: string`,
	}, nil
}

// ListenAndServe starts server
func (a *app) ListenAndServe(handler http.Handler) (*http.Server, <-chan error) {
	versionHandler := versionHandler()
	healthHandler := a.healthHandler()
	defaultHandler := ChainMiddlewares(handler, a.middlewares...)

	httpServer := &http.Server{
		Addr: a.listenAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				healthHandler.ServeHTTP(w, r)

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
	server, err := a.ListenAndServe(handler)
	a.WaitForTermination(err)

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("unable to shutdown HTTP server: %s", err)
	}
}

// WaitForTermination wait for error or SIGTERM/SIGINT signal
func (a *app) WaitForTermination(err <-chan error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	select {
	case err := <-err:
		logger.Error("%s", err)
	case signal := <-signals:
		logger.Info("%s received", signal)
		a.shutdown = true

		if a.graceDuration != 0 {
			logger.Info("Waiting %s for graceful shutdown", a.graceDuration)
			time.Sleep(a.graceDuration)
		}
	}
}

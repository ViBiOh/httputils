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
)

// App of package
type App interface {
	ListenAndServe(http.Handler, []model.Middleware, ...model.Pinger)
}

// Config of package
type Config struct {
	address  *string
	port     *uint
	cert     *string
	key      *string
	okStatus *int

	readTimeout     *string
	writeTimeout    *string
	idleTimeout     *string
	graceDuration   *string
	shutdownTimeout *string
}

type app struct {
	listenAddress string
	cert          string
	key           string
	okStatus      int

	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	graceDuration   time.Duration
	shutdownTimeout time.Duration
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		address:         flags.New(prefix, "http").Name("Address").Default("").Label("Listen address").ToString(fs),
		port:            flags.New(prefix, "http").Name("Port").Default(1080).Label("Listen port").ToUint(fs),
		cert:            flags.New(prefix, "http").Name("Cert").Default("").Label("Certificate file").ToString(fs),
		key:             flags.New(prefix, "http").Name("Key").Default("").Label("Key file").ToString(fs),
		okStatus:        flags.New(prefix, "http").Name("OkStatus").Default(http.StatusNoContent).Label("Healthy HTTP Status code").ToInt(fs),
		readTimeout:     flags.New(prefix, "http").Name("ReadTimeout").Default("5s").Label("Read Timeout").ToString(fs),
		writeTimeout:    flags.New(prefix, "http").Name("WriteTimeout").Default("10s").Label("Write Timeout").ToString(fs),
		idleTimeout:     flags.New(prefix, "http").Name("IdleTimeout").Default("2m").Label("Idle Timeout").ToString(fs),
		graceDuration:   flags.New(prefix, "http").Name("GraceDuration").Default("30s").Label("Grace duration when SIGTERM received").ToString(fs),
		shutdownTimeout: flags.New(prefix, "http").Name("ShutdownTimeout").Default("10s").Label("Shutdown Timeout").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return app{
		listenAddress: fmt.Sprintf("%s:%d", strings.TrimSpace(*config.address), *config.port),
		cert:          strings.TrimSpace(*config.cert),
		key:           strings.TrimSpace(*config.key),
		okStatus:      *config.okStatus,

		readTimeout:     safeParseDuration("ReadTimeout", *config.readTimeout, 5*time.Second),
		writeTimeout:    safeParseDuration("WriteTimeout", *config.writeTimeout, 10*time.Second),
		idleTimeout:     safeParseDuration("IdleTimeout", *config.idleTimeout, 2*time.Minute),
		graceDuration:   safeParseDuration("GraceDuration", *config.graceDuration, 30*time.Second),
		shutdownTimeout: safeParseDuration("ShutdownTimeout", *config.shutdownTimeout, 10*time.Second),
	}
}

// ListenAndServe starts server
func (a app) ListenAndServe(handler http.Handler, middlewares []model.Middleware, pingers ...model.Pinger) {
	versionHandler := versionHandler()
	defaultHandler := ChainMiddlewares(handler, middlewares...)

	done := make(chan struct{})
	healthHandler := healthHandler(a.okStatus, done, pingers...)

	httpServer := http.Server{
		Addr:         a.listenAddress,
		ReadTimeout:  a.readTimeout,
		WriteTimeout: a.writeTimeout,
		IdleTimeout:  a.idleTimeout,
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
		if len(a.cert) != 0 && len(a.key) != 0 {
			logger.Info("Listening with TLS")
			err <- httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			logger.Warn("Listening without TLS")
			err <- httpServer.ListenAndServe()
		}
	}()

	a.waitForGracefulShutdown(err, done)

	ctx, cancelFn := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancelFn()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("unable to shutdown HTTP server: %s", err)
	}
}

func (a app) waitForGracefulShutdown(err <-chan error, done chan<- struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	select {
	case err := <-err:
		logger.Error("%s", err)
	case sig := <-signals:
		logger.Info("%s received", sig)
		close(done)

		if a.graceDuration != 0 {
			logger.Info("Waiting %s for graceful shutdown", a.graceDuration)
			time.Sleep(a.graceDuration)
		}
	}
}

func versionHandler() http.Handler {
	versionValue := os.Getenv("VERSION")
	if len(versionValue) == 0 {
		versionValue = "development"
	}
	version := []byte(versionValue)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if _, err := w.Write(version); err != nil {
			logger.Error("%s", err)
		}
	})
}

func healthHandler(okStatus int, done <-chan struct{}, pingers ...model.Pinger) http.Handler {
	isShutdown := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if isShutdown() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		for _, pinger := range pingers {
			if err := pinger(); err != nil {
				logger.Error("unable to ping: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		w.WriteHeader(okStatus)
	})
}

// ChainMiddlewares chain middlewares call from last to first (so first item is the first called)
func ChainMiddlewares(handler http.Handler, middlewares ...model.Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}

	return result
}

func safeParseDuration(name string, value string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		logger.Warn("invalid %s value `%s`: %s", name, value, err)
		return defaultDuration
	}

	return duration
}

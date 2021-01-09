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
	ListenAndServe(http.Handler, []model.Pinger, ...model.Middleware)
	GetDone() <-chan struct{}
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
	done chan struct{}

	listenAddress string
	cert          string
	key           string

	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	graceDuration   time.Duration
	shutdownTimeout time.Duration

	okStatus int
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		address:         flags.New(prefix, "http").Name("Address").Default(flags.Default("Address", "", overrides)).Label("Listen address").ToString(fs),
		port:            flags.New(prefix, "http").Name("Port").Default(flags.Default("Port", 1080, overrides)).Label("Listen port").ToUint(fs),
		cert:            flags.New(prefix, "http").Name("Cert").Default(flags.Default("Cert", "", overrides)).Label("Certificate file").ToString(fs),
		key:             flags.New(prefix, "http").Name("Key").Default(flags.Default("Key", "", overrides)).Label("Key file").ToString(fs),
		okStatus:        flags.New(prefix, "http").Name("OkStatus").Default(flags.Default("OkStatus", http.StatusNoContent, overrides)).Label("Healthy HTTP Status code").ToInt(fs),
		readTimeout:     flags.New(prefix, "http").Name("ReadTimeout").Default(flags.Default("ReadTimeout", "5s", overrides)).Label("Read Timeout").ToString(fs),
		writeTimeout:    flags.New(prefix, "http").Name("WriteTimeout").Default(flags.Default("WriteTimeout", "10s", overrides)).Label("Write Timeout").ToString(fs),
		idleTimeout:     flags.New(prefix, "http").Name("IdleTimeout").Default(flags.Default("IdleTimeout", "2m", overrides)).Label("Idle Timeout").ToString(fs),
		graceDuration:   flags.New(prefix, "http").Name("GraceDuration").Default(flags.Default("GraceDuration", "30s", overrides)).Label("Grace duration when SIGTERM received").ToString(fs),
		shutdownTimeout: flags.New(prefix, "http").Name("ShutdownTimeout").Default(flags.Default("ShutdownTimeout", "10s", overrides)).Label("Shutdown Timeout").ToString(fs),
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

		done: make(chan struct{}),
	}
}

// ListenAndServe starts server
func (a app) GetDone() <-chan struct{} {
	return a.done
}

// ListenAndServe starts server
func (a app) ListenAndServe(handler http.Handler, pingers []model.Pinger, middlewares ...model.Middleware) {
	versionHandler := versionHandler()
	defaultHandler := ChainMiddlewares(handler, middlewares...)

	healthHandler := healthHandler(a.okStatus, a.done, pingers...)

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

	errors := make(chan error, 1)
	defer close(errors)

	go func() {
		if len(a.cert) != 0 && len(a.key) != 0 {
			logger.Info("Listening with TLS")
			errors <- httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			logger.Warn("Listening without TLS")
			errors <- httpServer.ListenAndServe()
		}
	}()

	a.waitForTermination(errors)
	if a.graceDuration != 0 {
		logger.Info("Waiting %s for graceful shutdown", a.graceDuration)
		time.Sleep(a.graceDuration)
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancelFn()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("unable to shutdown HTTP server: %s", err)
	}
}

func (a app) waitForTermination(errors <-chan error) {
	defer close(a.done)

	if err := WaitForTermination(errors); err != nil {
		logger.Error("%s", err)
		return
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

// WaitForTermination waits for an error or a n end signal
func WaitForTermination(errors <-chan error) error {
	signals := make(chan os.Signal, 1)
	defer close(signals)

	signal.Notify(signals, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-errors:
		return err
	case sig := <-signals:
		logger.Info("%s received", sig)
		return nil
	}
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

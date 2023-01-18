package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type App struct {
	done chan struct{}

	listenAddress string
	cert          string
	key           string

	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

type Config struct {
	address *string
	port    *uint
	cert    *string
	key     *string

	readTimeout     *time.Duration
	writeTimeout    *time.Duration
	idleTimeout     *time.Duration
	shutdownTimeout *time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		address:         flags.String(fs, prefix, "server", "Address", "Listen address", "", overrides),
		port:            flags.Uint(fs, prefix, "server", "Port", "Listen port (0 to disable)", 1080, overrides),
		cert:            flags.String(fs, prefix, "server", "Cert", "Certificate file", "", overrides),
		key:             flags.String(fs, prefix, "server", "Key", "Key file", "", overrides),
		readTimeout:     flags.Duration(fs, prefix, "server", "ReadTimeout", "Read Timeout", 5*time.Second, overrides),
		writeTimeout:    flags.Duration(fs, prefix, "server", "WriteTimeout", "Write Timeout", 10*time.Second, overrides),
		idleTimeout:     flags.Duration(fs, prefix, "server", "IdleTimeout", "Idle Timeout", 2*time.Minute, overrides),
		shutdownTimeout: flags.Duration(fs, prefix, "server", "ShutdownTimeout", "Shutdown Timeout", 10*time.Second, overrides),
	}
}

func New(config Config) App {
	port := *config.port
	done := make(chan struct{})

	if port == 0 {
		return App{
			done: done,
		}
	}

	return App{
		listenAddress: fmt.Sprintf("%s:%d", *config.address, port),
		cert:          *config.cert,
		key:           *config.key,

		readTimeout:     *config.readTimeout,
		writeTimeout:    *config.writeTimeout,
		idleTimeout:     *config.idleTimeout,
		shutdownTimeout: *config.shutdownTimeout,

		done: done,
	}
}

func (a App) Done() <-chan struct{} {
	return a.done
}

func (a App) Start(ctx context.Context, name string, handler http.Handler) {
	defer close(a.done)
	serverLogger := logger.WithField("server", name)

	if len(a.listenAddress) == 0 {
		serverLogger.Warn("No listen address")

		return
	}

	httpServer := http.Server{
		Addr:         a.listenAddress,
		ReadTimeout:  a.readTimeout,
		WriteTimeout: a.writeTimeout,
		IdleTimeout:  a.idleTimeout,
		Handler:      handler,
	}

	serverDone := make(chan struct{})

	go func() {
		defer close(serverDone)

		var err error
		if len(a.cert) != 0 && len(a.key) != 0 {
			serverLogger.Info("Listening on %s with TLS", a.listenAddress)
			err = httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			serverLogger.Warn("Listening on %s without TLS", a.listenAddress)
			err = httpServer.ListenAndServe()
		}

		if !errors.Is(err, http.ErrServerClosed) {
			serverLogger.Error("Server error: %s", err)
		}
	}()

	select {
	case <-ctx.Done():
	case <-serverDone:
	}

	ctx, cancelFn := context.WithTimeout(ctx, a.shutdownTimeout)
	defer cancelFn()

	serverLogger.Info("Server is shutting down")
	if err := httpServer.Shutdown(ctx); err != nil {
		serverLogger.Error("shutdown server: %s", err)
	}
}

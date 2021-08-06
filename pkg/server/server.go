package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

// App of package
type App interface {
	Start(string, <-chan struct{}, http.Handler)
	Done() <-chan struct{}
}

// Config of package
type Config struct {
	address *string
	port    *uint
	cert    *string
	key     *string

	readTimeout     *string
	writeTimeout    *string
	idleTimeout     *string
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
	shutdownTimeout time.Duration
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		address:         flags.New(prefix, "server").Name("Address").Default(flags.Default("Address", "", overrides)).Label("Listen address").ToString(fs),
		port:            flags.New(prefix, "server").Name("Port").Default(flags.Default("Port", 1080, overrides)).Label("Listen port (0 to disable)").ToUint(fs),
		cert:            flags.New(prefix, "server").Name("Cert").Default(flags.Default("Cert", "", overrides)).Label("Certificate file").ToString(fs),
		key:             flags.New(prefix, "server").Name("Key").Default(flags.Default("Key", "", overrides)).Label("Key file").ToString(fs),
		readTimeout:     flags.New(prefix, "server").Name("ReadTimeout").Default(flags.Default("ReadTimeout", "5s", overrides)).Label("Read Timeout").ToString(fs),
		writeTimeout:    flags.New(prefix, "server").Name("WriteTimeout").Default(flags.Default("WriteTimeout", "10s", overrides)).Label("Write Timeout").ToString(fs),
		idleTimeout:     flags.New(prefix, "server").Name("IdleTimeout").Default(flags.Default("IdleTimeout", "2m", overrides)).Label("Idle Timeout").ToString(fs),
		shutdownTimeout: flags.New(prefix, "server").Name("ShutdownTimeout").Default(flags.Default("ShutdownTimeout", "10s", overrides)).Label("Shutdown Timeout").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	port := *config.port
	done := make(chan struct{})

	if port == 0 {
		return app{
			done: done,
		}
	}

	return app{
		listenAddress: fmt.Sprintf("%s:%d", strings.TrimSpace(*config.address), port),
		cert:          strings.TrimSpace(*config.cert),
		key:           strings.TrimSpace(*config.key),

		readTimeout:     model.SafeParseDuration("ReadTimeout", *config.readTimeout, 5*time.Second),
		writeTimeout:    model.SafeParseDuration("WriteTimeout", *config.writeTimeout, 10*time.Second),
		idleTimeout:     model.SafeParseDuration("IdleTimeout", *config.idleTimeout, 2*time.Minute),
		shutdownTimeout: model.SafeParseDuration("ShutdownTimeout", *config.shutdownTimeout, 10*time.Second),

		done: done,
	}
}

// Done returns the chan closed when server is shutdown
func (a app) Done() <-chan struct{} {
	return a.done
}

func (a app) Start(name string, done <-chan struct{}, handler http.Handler) {
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

		if err != http.ErrServerClosed {
			serverLogger.Error("[%s] Server error: %s", name, err)
		}
	}()

	select {
	case <-done:
	case <-serverDone:
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancelFn()

	serverLogger.Info("Server is shutting down")
	if err := httpServer.Shutdown(ctx); err != nil {
		serverLogger.Error("unable to shutdown server: %s", err)
	}
}

// GracefulWait wait for all done chan to be closed
func GracefulWait(dones ...<-chan struct{}) {
	for _, done := range dones {
		<-done
	}
}

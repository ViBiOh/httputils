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
		port:            flags.New(prefix, "server").Name("Port").Default(flags.Default("Port", 1080, overrides)).Label("Listen port").ToUint(fs),
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
	return app{
		listenAddress: fmt.Sprintf("%s:%d", strings.TrimSpace(*config.address), *config.port),
		cert:          strings.TrimSpace(*config.cert),
		key:           strings.TrimSpace(*config.key),

		readTimeout:     model.SafeParseDuration("ReadTimeout", *config.readTimeout, 5*time.Second),
		writeTimeout:    model.SafeParseDuration("WriteTimeout", *config.writeTimeout, 10*time.Second),
		idleTimeout:     model.SafeParseDuration("IdleTimeout", *config.idleTimeout, 2*time.Minute),
		shutdownTimeout: model.SafeParseDuration("ShutdownTimeout", *config.shutdownTimeout, 10*time.Second),

		done: make(chan struct{}),
	}
}

// Done returns the chan closed when server is shutdown
func (a app) Done() <-chan struct{} {
	return a.done
}

func (a app) Start(name string, done <-chan struct{}, handler http.Handler) {
	defer close(a.done)

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
			logger.Info("[%s] Listening on %s with TLS", name, a.listenAddress)
			err = httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			logger.Warn("[%s] Listening on %s without TLS", name, a.listenAddress)
			err = httpServer.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			logger.Error("[%s] Server error: %s", name, err)
		}
	}()

	select {
	case <-done:
	case <-serverDone:
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancelFn()

	logger.Info("[%s] Server is shutting down", name)
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("[%s] unable to shutdown server: %s", name, err)
	}
}

// GracefulWait wait for all done chan to be closed
func GracefulWait(dones ...<-chan struct{}) {
	for _, done := range dones {
		<-done
	}
}

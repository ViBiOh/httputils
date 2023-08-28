package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/flags"
)

type Server struct {
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
	Address         string
	Cert            string
	Key             string
	Port            uint
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Address", "Listen address").Prefix(prefix).DocPrefix("server").StringVar(fs, &config.Address, "", overrides)
	flags.New("Port", "Listen port (0 to disable)").Prefix(prefix).DocPrefix("server").UintVar(fs, &config.Port, 1080, overrides)
	flags.New("Cert", "Certificate file").Prefix(prefix).DocPrefix("server").StringVar(fs, &config.Cert, "", overrides)
	flags.New("Key", "Key file").Prefix(prefix).DocPrefix("server").StringVar(fs, &config.Key, "", overrides)
	flags.New("ReadTimeout", "Read Timeout").Prefix(prefix).DocPrefix("server").DurationVar(fs, &config.ReadTimeout, 5*time.Second, overrides)
	flags.New("WriteTimeout", "Write Timeout").Prefix(prefix).DocPrefix("server").DurationVar(fs, &config.WriteTimeout, 10*time.Second, overrides)
	flags.New("IdleTimeout", "Idle Timeout").Prefix(prefix).DocPrefix("server").DurationVar(fs, &config.IdleTimeout, 2*time.Minute, overrides)
	flags.New("ShutdownTimeout", "Shutdown Timeout").Prefix(prefix).DocPrefix("server").DurationVar(fs, &config.ShutdownTimeout, 10*time.Second, overrides)

	return &config
}

func New(config *Config) Server {
	port := config.Port
	done := make(chan struct{})

	if port == 0 {
		return Server{
			done: done,
		}
	}

	return Server{
		listenAddress: fmt.Sprintf("%s:%d", config.Address, port),
		cert:          config.Cert,
		key:           config.Key,

		readTimeout:     config.ReadTimeout,
		writeTimeout:    config.WriteTimeout,
		idleTimeout:     config.IdleTimeout,
		shutdownTimeout: config.ShutdownTimeout,

		done: done,
	}
}

func (a Server) Done() <-chan struct{} {
	return a.done
}

func (a Server) Start(ctx context.Context, name string, handler http.Handler) {
	defer close(a.done)
	serverLogger := slog.With("server", name)

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
			serverLogger.Info("Listening with TLS", "address", a.listenAddress)
			err = httpServer.ListenAndServeTLS(a.cert, a.key)
		} else {
			serverLogger.Warn("Listening without TLS", "address", a.listenAddress)
			err = httpServer.ListenAndServe()
		}

		if !errors.Is(err, http.ErrServerClosed) {
			serverLogger.Error("Server error", "err", err)
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
		serverLogger.Error("shutdown server", "err", err)
	}
}

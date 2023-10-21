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
	done            chan struct{}
	logger          *slog.Logger
	cert            string
	key             string
	server          http.Server
	shutdownTimeout time.Duration
}

type Config struct {
	Address         string
	Name            string
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
	flags.New("Name", "Name").Prefix(prefix).DocPrefix("server").StringVar(fs, &config.Name, "http", overrides)
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
		done: done,

		cert:            config.Cert,
		key:             config.Key,
		shutdownTimeout: config.ShutdownTimeout,

		logger: slog.With("name", config.Name),
		server: http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Address, port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

func (a *Server) Done() <-chan struct{} {
	return a.done
}

func (a *Server) Start(ctx context.Context, name string, handler http.Handler) {
	defer close(a.done)

	if len(a.server.Addr) == 0 {
		a.logger.Warn("No listen address")

		return
	}

	var err error
	if len(a.cert) != 0 && len(a.key) != 0 {
		a.logger.Info("Listening with TLS", "address", a.server.Addr)
		err = a.server.ListenAndServeTLS(a.cert, a.key)
	} else {
		a.logger.Warn("Listening without TLS", "address", a.server.Addr)
		err = a.server.ListenAndServe()
	}

	if !errors.Is(err, http.ErrServerClosed) {
		a.logger.Error("Server error", "err", err)
	}
}

func (a *Server) Stop(ctx context.Context) {
	ctx, cancelFn := context.WithTimeout(ctx, a.shutdownTimeout)
	defer cancelFn()

	a.logger.Info("Server is shutting down")
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown server", "err", err)
	}
}

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
	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
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

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) Start(ctx context.Context, handler http.Handler) {
	defer close(s.done)

	if len(s.server.Addr) == 0 {
		s.logger.WarnContext(ctx, "No listen address")

		return
	}

	s.server.Handler = handler

	go func() {
		<-ctx.Done()
		s.Stop(cntxt.WithoutDeadline(ctx))
	}()

	var err error
	if len(s.cert) != 0 && len(s.key) != 0 {
		s.logger.InfoContext(ctx, "Listening with TLS", "address", s.server.Addr)
		err = s.server.ListenAndServeTLS(s.cert, s.key)
	} else {
		s.logger.WarnContext(ctx, "Listening without TLS", "address", s.server.Addr)
		err = s.server.ListenAndServe()
	}

	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.ErrorContext(ctx, "Server error", "err", err)
	}
}

func (s *Server) Stop(ctx context.Context) {
	ctx, cancelFn := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancelFn()

	s.logger.InfoContext(ctx, "Server is shutting down")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.ErrorContext(ctx, "shutdown server", "err", err)
	}
}

package health

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	LivePath  = "/health"
	ReadyPath = "/ready"
)

type Service struct {
	done chan struct{}
	end  chan struct{}

	pingers []model.Pinger

	okStatus      int
	graceDuration time.Duration
}

type Config struct {
	OkStatus      int
	GraceDuration time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("OkStatus", "Healthy HTTP Status code").Prefix(prefix).DocPrefix("http").IntVar(fs, &config.OkStatus, http.StatusNoContent, overrides)
	flags.New("GraceDuration", "Grace duration when signal received").Prefix(prefix).DocPrefix("http").DurationVar(fs, &config.GraceDuration, 30*time.Second, overrides)

	return &config
}

func New(config *Config, pingers ...model.Pinger) *Service {
	return &Service{
		okStatus:      config.OkStatus,
		graceDuration: config.GraceDuration,
		pingers:       pingers,

		done: make(chan struct{}),
		end:  make(chan struct{}),
	}
}

func (s *Service) Done(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		<-s.done
	}()

	return ctx
}

func (s *Service) End(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		<-s.end
	}()

	return ctx
}

func (s *Service) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.okStatus)
	})
}

func (s *Service) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-s.done:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			if s.isReady(r.Context()) {
				w.WriteHeader(s.okStatus)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
	})
}

func (s *Service) WaitForTermination(done <-chan struct{}, signals ...os.Signal) {
	defer close(s.end)

	s.waitForDone(done, signals...)

	select {
	case <-done:
	default:
		if s.graceDuration != 0 {
			slog.Info("Waiting for graceful shutdown", "duration", s.graceDuration)
			time.Sleep(s.graceDuration)
		}
	}
}

func (s *Service) waitForDone(done <-chan struct{}, signals ...os.Signal) {
	signalsChan := make(chan os.Signal, 1)
	defer close(signalsChan)

	signal.Notify(signalsChan, signals...)
	defer signal.Stop(signalsChan)

	defer close(s.done)

	select {
	case <-done:
	case sig := <-signalsChan:
		slog.Info("Signal received", "signal", sig.String())
	}
}

func (s *Service) isReady(ctx context.Context) bool {
	for _, pinger := range s.pingers {
		if err := pinger(ctx); err != nil {
			slog.Error("ping", "err", err)

			return false
		}
	}

	return true
}

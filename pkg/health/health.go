package health

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("OkStatus", "Healthy HTTP Status code").Prefix(prefix).DocPrefix("http").IntVar(fs, &config.OkStatus, http.StatusNoContent, overrides)
	flags.New("GraceDuration", "Grace duration when SIGTERM received").Prefix(prefix).DocPrefix("http").DurationVar(fs, &config.GraceDuration, 30*time.Second, overrides)

	return config
}

func New(config Config, pingers ...model.Pinger) *Service {
	return &Service{
		okStatus:      config.OkStatus,
		graceDuration: config.GraceDuration,
		pingers:       pingers,

		done: make(chan struct{}),
		end:  make(chan struct{}),
	}
}

func (a *Service) Done(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		<-a.done
	}()

	return ctx
}

func (a *Service) End(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		<-a.end
	}()

	return ctx
}

func (a *Service) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(a.okStatus)
	})
}

func (a *Service) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-a.done:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			if a.isReady(r.Context()) {
				w.WriteHeader(a.okStatus)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
	})
}

func (a *Service) WaitForTermination(done <-chan struct{}) {
	defer close(a.end)

	a.waitForDone(done, syscall.SIGTERM)

	select {
	case <-done:
	default:
		if a.graceDuration != 0 {
			slog.Info("Waiting for graceful shutdown", "duration", a.graceDuration)
			time.Sleep(a.graceDuration)
		}
	}
}

func (a *Service) waitForDone(done <-chan struct{}, signals ...os.Signal) {
	signalsChan := make(chan os.Signal, 1)
	defer close(signalsChan)

	signal.Notify(signalsChan, signals...)
	defer signal.Stop(signalsChan)

	defer close(a.done)

	select {
	case <-done:
	case sig := <-signalsChan:
		slog.Info("%s received", sig)
	}
}

func (a *Service) isReady(ctx context.Context) bool {
	for _, pinger := range a.pingers {
		if err := pinger(ctx); err != nil {
			slog.Error("ping", "err", err)

			return false
		}
	}

	return true
}

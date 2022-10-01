package health

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	LivePath  = "/health"
	ReadyPath = "/ready"
)

type App struct {
	done chan struct{}
	end  chan struct{}

	pingers []model.Pinger

	okStatus      int
	graceDuration time.Duration
}

type Config struct {
	okStatus      *int
	graceDuration *time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		okStatus:      flags.Int(fs, prefix, "http", "OkStatus", "Healthy HTTP Status code", http.StatusNoContent, overrides),
		graceDuration: flags.Duration(fs, prefix, "http", "GraceDuration", "Grace duration when SIGTERM received", 30*time.Second, overrides),
	}
}

func New(config Config, pingers ...model.Pinger) App {
	return App{
		okStatus:      *config.okStatus,
		graceDuration: *config.graceDuration,
		pingers:       pingers,

		done: make(chan struct{}),
		end:  make(chan struct{}),
	}
}

func (a App) Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		<-a.done
	}()

	return ctx
}

func (a App) Done() <-chan struct{} {
	return a.done
}

func (a App) End() <-chan struct{} {
	return a.end
}

func (a App) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(a.okStatus)
	})
}

func (a App) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-a.done:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			if a.isReady() {
				w.WriteHeader(a.okStatus)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
	})
}

func (a App) WaitForTermination(done <-chan struct{}) {
	defer close(a.end)

	a.waitForDone(done, syscall.SIGTERM)

	select {
	case <-done:
	default:
		if a.graceDuration != 0 {
			logger.Info("Waiting %s for graceful shutdown", a.graceDuration)
			time.Sleep(a.graceDuration)
		}
	}
}

func (a App) waitForDone(done <-chan struct{}, signals ...os.Signal) {
	signalsChan := make(chan os.Signal, 1)
	defer close(signalsChan)

	signal.Notify(signalsChan, signals...)
	defer signal.Stop(signalsChan)

	defer close(a.done)

	select {
	case <-done:
	case sig := <-signalsChan:
		logger.Info("%s received", sig)
	}
}

func (a App) isReady() bool {
	for _, pinger := range a.pingers {
		if err := pinger(); err != nil {
			logger.Error("ping: %s", err)

			return false
		}
	}

	return true
}

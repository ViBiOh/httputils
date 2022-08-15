package health

import (
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
	// LivePath is the path for checking that HTTP service is live
	LivePath = "/health"

	// ReadyPath is the path for checking that HTTP service is ready (checking dependencies)
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

// Done returns the chan closed when SIGTERM is received
func (a App) Done() <-chan struct{} {
	return a.done
}

// End returns the chan closed when graceful duration is over
func (a App) End() <-chan struct{} {
	return a.end
}

// HealthHandler for request. Should be use with net/http
func (a App) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(a.okStatus)
	})
}

// ReadyHandler for request. Should be use with net/http
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

// WaitForTermination waits for SIGTERM or done plus grace duration
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

// waitForDone waits for the SIGTERM signal or close of done
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

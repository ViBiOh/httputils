package health

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	// HealthPath is the path for checking that HTTP service is live
	HealthPath = "/health"

	// ReadyPath is the path for checking that HTTP service is ready (checking dependencies)
	ReadyPath = "/ready"
)

// App of package
type App interface {
	Handler() http.Handler
	WaitForTermination(<-chan struct{})
	Done() <-chan struct{}
	End() <-chan struct{}
}

// Config of package
type Config struct {
	okStatus      *int
	graceDuration *string
}

type app struct {
	done chan struct{}
	end  chan struct{}

	pingers []model.Pinger

	okStatus      int
	graceDuration time.Duration
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		okStatus:      flags.New(prefix, "http").Name("OkStatus").Default(flags.Default("OkStatus", http.StatusNoContent, overrides)).Label("Healthy HTTP Status code").ToInt(fs),
		graceDuration: flags.New(prefix, "http").Name("GraceDuration").Default(flags.Default("GraceDuration", "30s", overrides)).Label("Grace duration when SIGTERM received").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, pingers ...model.Pinger) App {
	return app{
		okStatus:      *config.okStatus,
		graceDuration: model.SafeParseDuration("GraceDuration", *config.graceDuration, 30*time.Second),
		pingers:       pingers,

		done: make(chan struct{}),
		end:  make(chan struct{}),
	}
}

// Done returns the chan closed when SIGTERM is received
func (a app) Done() <-chan struct{} {
	return a.done
}

// End returns the chan closed when graceful duration is over
func (a app) End() <-chan struct{} {
	return a.end
}

func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if a.isShutdown() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		if r.URL.Path == HealthPath {
			w.WriteHeader(a.okStatus)
			return
		}

		for _, pinger := range a.pingers {
			if err := pinger(); err != nil {
				logger.Error("unable to ping: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		w.WriteHeader(a.okStatus)
	})
}

// WaitForTermination waits for the SIGTERM signal or close of done
func (a app) WaitForTermination(done <-chan struct{}) {
	defer close(a.end)

	signals := make(chan os.Signal, 1)
	defer close(signals)

	signal.Notify(signals, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case <-done:
		close(a.done)
		return
	case sig := <-signals:
		close(a.done)
		logger.Info("%s received", sig)
	}

	if a.graceDuration != 0 {
		logger.Info("Waiting %s for graceful shutdown", a.graceDuration)
		time.Sleep(a.graceDuration)
	}
}

func (a app) isShutdown() bool {
	select {
	case <-a.done:
		return true
	default:
		return false
	}
}

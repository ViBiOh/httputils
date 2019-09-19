package scheduler

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
)

var (
	// ErrRetryCanceled cancel retry loop
	ErrRetryCanceled = errors.New("retry canceled")
)

// Config of package
type Config struct {
	onStart  *bool
	hour     *int
	minute   *int
	interval *string
	retry    *string
	maxRetry *uint
	timezone *string
}

// App of package
type App interface {
	Start()
}

type app struct {
	onStart  bool
	hour     int
	minute   int
	location *time.Location

	interval time.Duration
	retry    time.Duration
	maxRetry uint

	task Task
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		onStart:  tools.NewFlag(prefix, "scheduler").Name("OnStart").Default(false).Label("Start scheduler on start").ToBool(fs),
		hour:     tools.NewFlag(prefix, "scheduler").Name("Hour").Default(8).Label("Hour of running").ToInt(fs),
		minute:   tools.NewFlag(prefix, "scheduler").Name("Minute").Default(0).Label("Minute of running").ToInt(fs),
		timezone: tools.NewFlag(prefix, "scheduler").Name("Timezone").Default("Europe/Paris").Label("Timezone of running").ToString(fs),
		interval: tools.NewFlag(prefix, "scheduler").Name("Interval").Default("24h").Label("Duration between two runs").ToString(fs),
		retry:    tools.NewFlag(prefix, "scheduler").Name("Retry").Default("10m").Label("Duration between two retries").ToString(fs),
		maxRetry: tools.NewFlag(prefix, "scheduler").Name("MaxRetry").Default(10).Label("Max retry").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, task Task) (App, error) {
	location, err := time.LoadLocation(strings.TrimSpace(*config.timezone))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	interval, err := time.ParseDuration(strings.TrimSpace(*config.interval))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	retry, err := time.ParseDuration(strings.TrimSpace(*config.retry))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return app{
		onStart:  *config.onStart,
		hour:     *config.hour,
		minute:   *config.minute,
		interval: interval,
		retry:    retry,
		maxRetry: *config.maxRetry,
		location: location,
		task:     task,
	}, nil
}

// Start scheduler
func (a app) Start() {
	if a.onStart {
		a.scheduleOnStart()
	}

	a.scheduleDaily()
}

func (a app) getNow() time.Time {
	return time.Now().In(a.location)
}

func (a app) scheduleOnStart() {
	timer := getTimer(a.getNow().Add(time.Second * 2))

	for {
		a.runIteration(timer)
		timer.Reset(a.interval)
	}
}

func (a app) scheduleDaily() {
	timer := getTimer(a.getNextDailyTick())

	for {
		a.runIteration(timer)
		timer = getTimer(a.getNextDailyTick())
	}
}

func (a app) runIteration(timer *time.Timer) {
	retryCount := uint(0)

	for {
		currentTime := <-timer.C
		ctx := context.Background()

		err := a.task.Do(ctx, currentTime)
		if err == nil {
			return
		}

		logger.Error("%#v", err)

		if err == ErrRetryCanceled {
			return
		}

		retryCount++
		if retryCount >= a.maxRetry {
			logger.Error("max retry exceeded")
			return
		}

		timer.Reset(a.retry)
		logger.Warn("Retrying in %s", a.retry)
	}
}

func (a app) getNextDailyTick() time.Time {
	currentTime := a.getNow()
	nextTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), a.hour, a.minute, 0, 0, a.location)

	if !nextTime.After(currentTime) {
		nextTime = nextTime.Add(a.interval)
	}

	return nextTime
}

func getTimer(tick time.Time) *time.Timer {
	logger.Info("Next run at %s", tick.String())

	return time.NewTimer(time.Until(tick))
}

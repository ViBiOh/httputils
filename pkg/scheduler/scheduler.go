package scheduler

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
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
	maxRetry *int
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
	maxRetry int

	task Task
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "scheduler"
	}

	return Config{
		onStart:  fs.Bool(tools.ToCamel(fmt.Sprintf("%sOnStart", prefix)), false, fmt.Sprintf("[%s] Start scheduler on start", docPrefix)),
		hour:     fs.Int(tools.ToCamel(fmt.Sprintf("%sHour", prefix)), 8, fmt.Sprintf("[%s] Hour of running", docPrefix)),
		minute:   fs.Int(tools.ToCamel(fmt.Sprintf("%sMinute", prefix)), 0, fmt.Sprintf("[%s] Minute of running", docPrefix)),
		timezone: fs.String(tools.ToCamel(fmt.Sprintf("%sTimezone", prefix)), "Europe/Paris", fmt.Sprintf("[%s] Timezone of running", docPrefix)),
		interval: fs.String(tools.ToCamel(fmt.Sprintf("%sInterval", prefix)), "24h", fmt.Sprintf("[%s] Duration between two runs", docPrefix)),
		retry:    fs.String(tools.ToCamel(fmt.Sprintf("%sRetry", prefix)), "10m", fmt.Sprintf("[%s] Duration between two retries", docPrefix)),
		maxRetry: fs.Int(tools.ToCamel(fmt.Sprintf("%sMaxRetry", prefix)), 10, fmt.Sprintf("[%s] Max retry", docPrefix)),
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
	timer := a.getTimer()
	retryCount := 0

	for {
		for {
			currentTime := <-timer.C
			ctx := context.Background()

			err := a.task.Do(ctx, currentTime)
			if err == nil {

			}

			logger.Error("%#v", err)

			if err == ErrRetryCanceled {
				break
			}

			retryCount++
			if retryCount >= a.maxRetry {
				logger.Error("max retry exceeded")
				break
			}

			timer.Reset(a.retry)
			logger.Warn("Retrying in %s", a.retry)
		}

		timer.Reset(a.interval)
	}
}

func (a app) getNextTick() (time.Time, time.Time) {
	currentTime := time.Now().In(a.location)

	if a.onStart {
		return currentTime.Add(time.Second * 5), currentTime
	}

	return time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), a.hour, a.minute, 0, 0, a.location), currentTime
}

func (a app) getTimer() *time.Timer {
	nextTime, currentTime := a.getNextTick()
	if !nextTime.After(currentTime) {
		nextTime = nextTime.Add(a.interval)
	}

	logger.Info("Next run at %s", nextTime.String())

	return time.NewTimer(time.Until(nextTime))
}

func (a app) scheduler() {
}

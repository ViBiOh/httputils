package cron

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

// Semaphore client
//
//go:generate mockgen -destination ../mocks/redis.go -mock_names Semaphore=Semaphore -package mocks github.com/ViBiOh/httputils/v4/pkg/cron Semaphore
type Semaphore interface {
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
}

const (
	hourFormat = "15:04"
)

var _ fmt.Stringer = New()

// Cron definition
type Cron struct {
	tracer       trace.Tracer
	clock        clock.Clock
	semaphoreApp Semaphore

	signal  os.Signal
	dayTime time.Time
	now     chan time.Time
	name    string

	onError func(error)
	errors  []error

	retryInterval time.Duration
	timeout       time.Duration
	interval      time.Duration

	maxRetry uint
	day      byte
}

// New creates new cron
func New() *Cron {
	return &Cron{
		dayTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		now:     make(chan time.Time, 1),
		onError: func(err error) {
			fmt.Println(err)
		},
	}
}

func (c *Cron) String() string {
	var buffer strings.Builder

	if c.interval != 0 {
		fmt.Fprintf(&buffer, "each: %s", c.interval)
	} else {
		fmt.Fprintf(&buffer, "day: %07b, at: %02d:%02d, in: %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.dayTime.Location())
	}

	fmt.Fprintf(&buffer, ", retry: %d times every %s", c.maxRetry, c.retryInterval)

	if c.semaphoreApp != nil {
		fmt.Fprintf(&buffer, ", in exclusive mode as `%s` with %s timeout", c.name, c.timeout)
	}

	for _, err := range c.errors {
		fmt.Fprintf(&buffer, ", error=`%s`", err)
	}

	return buffer.String()
}

// Days sets recurence to every day
func (c *Cron) Days() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday().Saturday().Sunday()
}

// Weekdays sets recurence to every day except sunday and saturday
func (c *Cron) Weekdays() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday()
}

// Sunday sets recurence to every Sunday
func (c *Cron) Sunday() *Cron {
	c.day = c.day | 1<<time.Sunday

	return c
}

// Monday sets recurence to every Monday
func (c *Cron) Monday() *Cron {
	c.day = c.day | 1<<time.Monday

	return c
}

// Tuesday sets recurence to every Tuesday
func (c *Cron) Tuesday() *Cron {
	c.day = c.day | 1<<time.Tuesday

	return c
}

// Wednesday sets recurence to every Wednesday
func (c *Cron) Wednesday() *Cron {
	c.day = c.day | 1<<time.Wednesday

	return c
}

// Thursday sets recurence to every Thursday
func (c *Cron) Thursday() *Cron {
	c.day = c.day | 1<<time.Thursday

	return c
}

// Friday sets recurence to every Friday
func (c *Cron) Friday() *Cron {
	c.day = c.day | 1<<time.Friday

	return c
}

// Saturday sets recurence to every Saturday
func (c *Cron) Saturday() *Cron {
	c.day = c.day | 1<<time.Saturday

	return c
}

// At sets hour of run in format HH:MM
func (c *Cron) At(hour string) *Cron {
	hourTime, err := time.ParseInLocation(hourFormat, hour, c.dayTime.Location())

	if err != nil {
		c.errors = append(c.errors, err)
	} else {
		c.dayTime = hourTime
	}

	return c
}

// Exclusive runs cron in an exclusive manner with a distributed lock on Redis
func (c *Cron) Exclusive(semaphoreApp Semaphore, name string, timeout time.Duration) *Cron {
	c.semaphoreApp = semaphoreApp
	c.name = name
	c.timeout = timeout

	return c
}

// In sets timezone
func (c *Cron) In(tz string) *Cron {
	timezone, err := time.LoadLocation(tz)
	if err != nil {
		c.errors = append(c.errors, err)
		return c
	}

	hourTime, err := time.ParseInLocation(hourFormat, fmt.Sprintf("%02d:%02d", c.dayTime.Hour(), c.dayTime.Minute()), timezone)
	if err != nil {
		c.errors = append(c.errors, err)
		return c
	}

	c.dayTime = hourTime

	return c
}

// Each sets interval of each run
func (c *Cron) Each(interval time.Duration) *Cron {
	if c.day != 0 {
		c.errors = append(c.errors, errors.New("cannot set interval and days on the same cron"))
	}

	c.interval = interval

	return c
}

// Retry sets interval retry if action failed
func (c *Cron) Retry(retryInterval time.Duration) *Cron {
	c.retryInterval = retryInterval

	return c
}

// MaxRetry sets maximum retry count
func (c *Cron) MaxRetry(maxRetry uint) *Cron {
	c.maxRetry = maxRetry

	return c
}

// OnSignal sets signal listened for trigerring cron
func (c *Cron) OnSignal(signal os.Signal) *Cron {
	c.signal = signal

	return c
}

// OnError defines error handling function
func (c *Cron) OnError(onError func(error)) *Cron {
	c.onError = onError

	return c
}

func (c *Cron) findMatchingDay(nextTime time.Time) time.Time {
	for day := nextTime.Weekday(); (1 << day & c.day) == 0; day = (day + 1) % 7 {
		nextTime = nextTime.AddDate(0, 0, 1)
	}

	return nextTime
}

func (c *Cron) getTickerDuration(shouldRetry bool) time.Duration {
	if shouldRetry && c.retryInterval != 0 {
		return c.retryInterval
	}

	if c.interval != 0 {
		return c.interval
	}

	tz := c.dayTime.Location()
	now := c.clock.Now().In(tz)

	nextTime := c.findMatchingDay(time.Date(now.Year(), now.Month(), now.Day(), c.dayTime.Hour(), c.dayTime.Minute(), 0, 0, tz))
	if nextTime.Before(now) {
		nextTime = c.findMatchingDay(nextTime.AddDate(0, 0, 1))
	}

	return nextTime.Sub(now)
}

func (c *Cron) hasError() bool {
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			c.onError(err)
		}
		return true
	}

	if c.day == 0 && c.interval == 0 {
		c.onError(errors.New("no schedule configuration"))
		return true
	}

	if c.maxRetry != 0 && c.retryInterval == 0 {
		c.onError(errors.New("no retry interval for max retry"))
		return true
	}

	return false
}

// WithTracer starts a span on each context
func (c *Cron) WithTracer(tracer trace.Tracer) *Cron {
	c.tracer = tracer
	return c
}

// Now run cron now
func (c *Cron) Now() *Cron {
	c.now <- c.clock.Now()

	return c
}

// Start cron
func (c *Cron) Start(action func(context.Context) error, done <-chan struct{}) {
	if c.hasError() {
		return
	}

	retryCount := uint(0)
	shouldRetry := false

	do := func(ctx context.Context) {
		if err := action(ctx); err != nil {
			c.onError(err)

			retryCount++
			shouldRetry = retryCount <= c.maxRetry
		} else {
			retryCount = 0
			shouldRetry = false
		}
	}

	run := func() {
		ctx, end := tracer.StartSpan(context.Background(), c.tracer, "cron")
		defer end()

		if c.semaphoreApp == nil {
			do(ctx)
			return
		}

		if _, err := c.semaphoreApp.Exclusive(ctx, c.name, c.timeout, func(ctx context.Context) error {
			do(ctx)
			return nil
		}); err != nil {
			c.onError(err)
		}
	}

	signals := make(chan os.Signal, 1)
	defer close(signals)

	if c.signal != nil {
		signal.Notify(signals, c.signal)
		defer signal.Stop(signals)
	}

	for {
		select {
		case <-done:
			return
		case <-signals:
			run()
		case <-time.After(c.getTickerDuration(shouldRetry)):
			run()
		case _, ok := <-c.now:
			if ok {
				run()
			}
		}
	}
}

// Shutdown cron, do not attempt Start() after
func (c *Cron) Shutdown() {
	close(c.now)
}

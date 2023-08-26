package cron

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cron.go -destination ../mocks/cron.go -package mocks -mock_names Semaphore=Semaphore

type Semaphore interface {
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
}

type GetNow func() time.Time

const hourFormat = "15:04"

var _ fmt.Stringer = New()

type Cron struct {
	tracer    trace.Tracer
	clock     GetNow
	semaphore Semaphore

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

func New() *Cron {
	return &Cron{
		dayTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		now:     make(chan time.Time, 1),
		clock:   time.Now,
		onError: func(err error) {
			fmt.Println(err)
		},
	}
}

func (c *Cron) String() string {
	var buffer strings.Builder

	if c.interval != 0 {
		_, _ = fmt.Fprintf(&buffer, "each: %s", c.interval)
	} else {
		_, _ = fmt.Fprintf(&buffer, "day: %07b, at: %02d:%02d, in: %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.dayTime.Location())
	}

	_, _ = fmt.Fprintf(&buffer, ", retry: %d times every %s", c.maxRetry, c.retryInterval)

	if c.semaphore != nil {
		_, _ = fmt.Fprintf(&buffer, ", in exclusive mode as `%s` with %s timeout", c.name, c.timeout)
	}

	for _, err := range c.errors {
		_, _ = fmt.Fprintf(&buffer, ", error=`%s`", err)
	}

	return buffer.String()
}

func (c *Cron) Days() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday().Saturday().Sunday()
}

func (c *Cron) Weekdays() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday()
}

func (c *Cron) Sunday() *Cron {
	c.day = c.day | 1<<time.Sunday

	return c
}

func (c *Cron) Monday() *Cron {
	c.day = c.day | 1<<time.Monday

	return c
}

func (c *Cron) Tuesday() *Cron {
	c.day = c.day | 1<<time.Tuesday

	return c
}

func (c *Cron) Wednesday() *Cron {
	c.day = c.day | 1<<time.Wednesday

	return c
}

func (c *Cron) Thursday() *Cron {
	c.day = c.day | 1<<time.Thursday

	return c
}

func (c *Cron) Friday() *Cron {
	c.day = c.day | 1<<time.Friday

	return c
}

func (c *Cron) Saturday() *Cron {
	c.day = c.day | 1<<time.Saturday

	return c
}

func (c *Cron) At(hour string) *Cron {
	hourTime, err := time.ParseInLocation(hourFormat, hour, c.dayTime.Location())

	if err != nil {
		c.errors = append(c.errors, err)
	} else {
		c.dayTime = hourTime
	}

	return c
}

func (c *Cron) Exclusive(semaphore Semaphore, name string, timeout time.Duration) *Cron {
	c.semaphore = semaphore
	c.name = name
	c.timeout = timeout

	return c
}

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

func (c *Cron) Each(interval time.Duration) *Cron {
	if c.day != 0 {
		c.errors = append(c.errors, errors.New("cannot set interval and days on the same cron"))
	}

	c.interval = interval

	return c
}

func (c *Cron) Retry(retryInterval time.Duration) *Cron {
	c.retryInterval = retryInterval

	return c
}

func (c *Cron) MaxRetry(maxRetry uint) *Cron {
	c.maxRetry = maxRetry

	return c
}

func (c *Cron) OnSignal(signal os.Signal) *Cron {
	c.signal = signal

	return c
}

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
	now := c.clock().In(tz)

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

func (c *Cron) WithTracerProvider(tracerProvider trace.TracerProvider) *Cron {
	if tracerProvider != nil {
		c.tracer = tracerProvider.Tracer("cron")
	}

	return c
}

func (c *Cron) Now() *Cron {
	c.now <- c.clock()

	return c
}

func (c *Cron) Start(ctx context.Context, action func(context.Context) error) {
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
		var err error

		ctx, end := telemetry.StartSpan(ctx, c.tracer, "cron")
		defer end(&err)

		if c.semaphore == nil {
			do(ctx)

			return
		}

		if _, err = c.semaphore.Exclusive(ctx, c.name, c.timeout, func(ctx context.Context) error {
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

	done := ctx.Done()

	for {
		if c.iterate(done, signals, shouldRetry, run) {
			return
		}
	}
}

func (c *Cron) iterate(done <-chan struct{}, signals <-chan os.Signal, shouldRetry bool, run func()) bool {
	var output bool

	timer := time.After(c.getTickerDuration(shouldRetry))
	doneTimer := make(chan struct{})
	go func() {
		defer close(doneTimer)
		<-timer
	}()

	select {
	case <-done:
		output = true
	case <-signals:
		run()
	case <-doneTimer:
		run()
	case _, ok := <-c.now:
		if ok {
			run()
		}
	}

	return output
}

func (c *Cron) Shutdown() {
	close(c.now)
}

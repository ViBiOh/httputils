package cron

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	hourFormat = "15:04"
)

var (
	_ fmt.Stringer = New()
)

// Clock give time
type Clock struct {
	now time.Time
}

// Now return current time
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

// Cron definition
type Cron struct {
	retryInterval time.Duration
	interval      time.Duration
	dayTime       time.Time

	clock  *Clock
	now    chan time.Time
	signal os.Signal

	onError func(error)
	errors  []error

	day      byte
	maxRetry uint
}

// New creates new cron
func New() *Cron {
	return &Cron{
		dayTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		now:     make(chan time.Time, 1),
		errors:  make([]error, 0),
		onError: func(err error) {
			fmt.Println(err)
		},
	}
}

func (c *Cron) String() string {
	var buffer strings.Builder

	if c.interval != 0 {
		buffer.WriteString(fmt.Sprintf("each: %s", c.interval))
	} else {
		buffer.WriteString(fmt.Sprintf("day: %07b, at: %02d:%02d, in: %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.dayTime.Location()))
	}

	buffer.WriteString(fmt.Sprintf(", retry: %d times every %s", c.maxRetry, c.retryInterval))

	for _, err := range c.errors {
		buffer.WriteString(fmt.Sprintf(", error=`%s`", err))
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

// Clock sets clock that give current time, mostly for testing purpose
func (c *Cron) Clock(clock *Clock) *Cron {
	c.clock = clock

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

// Now run cron now
func (c *Cron) Now() *Cron {
	c.now <- c.clock.Now()

	return c
}

// Start cron
func (c *Cron) Start(action func(time.Time) error, done <-chan struct{}) {
	if c.hasError() {
		return
	}

	retryCount := uint(0)
	shouldRetry := false

	do := func(now time.Time) {
		if err := action(now); err != nil {
			c.onError(err)

			retryCount++
			shouldRetry = retryCount <= c.maxRetry
		} else {
			retryCount = 0
			shouldRetry = false
		}
	}

	for {
		ticker := time.NewTicker(c.getTickerDuration(shouldRetry))
		defer ticker.Stop()

		signals := make(chan os.Signal, 1)
		defer close(signals)

		if c.signal != nil {
			signal.Notify(signals, c.signal)
			defer signal.Stop(signals)
		}

		select {
		case <-done:
			return
		case <-signals:
			do(c.clock.Now())
		case now := <-ticker.C:
			do(now)
		case now, ok := <-c.now:
			if ok {
				do(now)
			}
		}
	}
}

// Shutdown cron, do not attempt Start() after
func (c *Cron) Shutdown() {
	close(c.now)
}

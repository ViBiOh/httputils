package cron

import (
	"errors"
	"fmt"
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
	day           byte
	dayTime       time.Time
	timezone      *time.Location
	interval      time.Duration
	maxRetry      uint
	retryInterval time.Duration
	onStart       bool

	now    chan time.Time
	errors []error

	clock *Clock
}

// New create new cron
func New() *Cron {
	return &Cron{
		dayTime:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
		timezone: time.Local,
		now:      make(chan time.Time, 1),
		errors:   make([]error, 0),
	}
}

func (c *Cron) String() string {
	var buffer strings.Builder

	if c.interval != 0 {
		buffer.WriteString(fmt.Sprintf("each: %s", c.interval))
	} else {
		buffer.WriteString(fmt.Sprintf("day: %07b, at: %02d:%02d, in: %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.timezone))
	}

	buffer.WriteString(fmt.Sprintf(", retry: %d times every %s", c.maxRetry, c.retryInterval))

	return buffer.String()
}

// Days set recurence to every day
func (c *Cron) Days() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday().Saturday().Sunday()
}

// Weekdays set recurence to every day except sunday and saturday
func (c *Cron) Weekdays() *Cron {
	return c.Monday().Tuesday().Wednesday().Thursday().Friday()
}

// Sunday set recurence to every Sunday
func (c *Cron) Sunday() *Cron {
	c.day = c.day | 1<<time.Sunday

	return c
}

// Monday set recurence to every Monday
func (c *Cron) Monday() *Cron {
	c.day = c.day | 1<<time.Monday

	return c
}

// Tuesday set recurence to every Tuesday
func (c *Cron) Tuesday() *Cron {
	c.day = c.day | 1<<time.Tuesday

	return c
}

// Wednesday set recurence to every Wednesday
func (c *Cron) Wednesday() *Cron {
	c.day = c.day | 1<<time.Wednesday

	return c
}

// Thursday set recurence to every Thursday
func (c *Cron) Thursday() *Cron {
	c.day = c.day | 1<<time.Thursday

	return c
}

// Friday set recurence to every Friday
func (c *Cron) Friday() *Cron {
	c.day = c.day | 1<<time.Friday

	return c
}

// Saturday set recurence to every Saturday
func (c *Cron) Saturday() *Cron {
	c.day = c.day | 1<<time.Saturday

	return c
}

// At set hour of run in format HH:MM
func (c *Cron) At(hour string) *Cron {
	hourTime, err := time.Parse(hourFormat, hour)

	if err != nil {
		c.errors = append(c.errors, err)
	} else {
		c.dayTime = hourTime
	}

	return c
}

// In set timezone
func (c *Cron) In(tz string) *Cron {
	timezone, err := time.LoadLocation(tz)
	if err != nil {
		c.errors = append(c.errors, err)
	} else {
		c.timezone = timezone
	}

	return c
}

// Each set interval of each run
func (c *Cron) Each(interval time.Duration) *Cron {
	if c.day != 0 {
		c.errors = append(c.errors, errors.New("cannot set interval and days on the same cron"))
	}

	c.interval = interval

	return c
}

// Retry set interval retry if action failed
func (c *Cron) Retry(retryInterval time.Duration) *Cron {
	c.retryInterval = retryInterval

	return c
}

// MaxRetry set maximum retry count
func (c *Cron) MaxRetry(maxRetry uint) *Cron {
	c.maxRetry = maxRetry

	return c
}

// Clock set clock that give current time, mostly for testing purpose
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

	now := c.clock.Now().In(c.timezone)

	nextTime := c.findMatchingDay(time.Date(now.Year(), now.Month(), now.Day(), c.dayTime.Hour(), c.dayTime.Minute(), 0, 0, c.timezone))
	if nextTime.Before(now) {
		nextTime = c.findMatchingDay(nextTime.AddDate(0, 0, 1))
	}

	return nextTime.Sub(now)
}

func (c *Cron) hasError(onError func(error)) bool {
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			onError(err)
		}
		return true
	}

	if c.day == 0 && c.interval == 0 {
		onError(errors.New("no schedule configuration"))
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
func (c *Cron) Start(action func(time.Time) error, onError func(error)) {
	if c.hasError(onError) {
		return
	}

	retryCount := uint(0)
	shouldRetry := false

	do := func(now time.Time) {
		if err := action(now); err != nil {
			onError(err)

			retryCount++
			shouldRetry = retryCount <= c.maxRetry
		} else {
			retryCount = 0
			shouldRetry = false
		}
	}

	for {
		duration := c.getTickerDuration(shouldRetry)
		ticker := time.NewTicker(duration)

		select {
		case now, ok := <-c.now:
			ticker.Stop()
			if ok {
				do(now)
			} else {
				return
			}
		case now := <-ticker.C:
			do(now)
		}
	}
}

// Stop cron
func (c *Cron) Stop() {
	close(c.now)
}

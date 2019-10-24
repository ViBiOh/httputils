package cron

import (
	"errors"
	"fmt"
	"time"
)

const (
	hourFormat = "15:04"
)

var (
	_ fmt.Stringer = NewCron()
)

// Cron definition
type Cron struct {
	day           byte
	dayTime       time.Time
	timezone      *time.Location
	interval      time.Duration
	maxRetry      uint
	retryInterval time.Duration

	done   chan struct{}
	errors []error
}

// NewCron create new cron
func NewCron() *Cron {
	return &Cron{
		dayTime:  time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
		timezone: time.Local,
		errors:   make([]error, 0),
	}
}

func (c *Cron) String() string {
	return fmt.Sprintf("day: %08b, at: %d:%d, each: %s, retry: %d every %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.interval, c.maxRetry, c.retryInterval)
}

// Days set recurence to every day
func (c *Cron) Days() *Cron {
	c.day = c.day | 0xFF

	return c
}

// Weekdays set recurence to every day except sunday and saturday
func (c *Cron) Weekdays() *Cron {
	c.Days()

	c.day = c.day ^ 1<<time.Sunday
	c.day = c.day ^ 1<<time.Saturday

	return c
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
func (c *Cron) In(timezone *time.Location) *Cron {
	c.timezone = timezone

	return c
}

// Each set interval of each run
func (c *Cron) Each(interval time.Duration) *Cron {
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

func (c *Cron) findMatchingDay(nextTime time.Time) time.Time {
	for day := nextTime.Weekday(); (1 << day & c.day) == 0; day = (day + 1) % 7 {
		nextTime = nextTime.AddDate(0, 0, 1)
	}

	return nextTime
}

func (c *Cron) getTicker(shouldRetry bool) *time.Ticker {
	if shouldRetry && c.retryInterval != 0 {
		return time.NewTicker(c.retryInterval)
	}

	if c.interval != 0 {
		return time.NewTicker(c.interval)
	}

	now := time.Now()

	nextTime := c.findMatchingDay(time.Date(now.Year(), now.Month(), now.Day(), c.dayTime.Hour(), c.dayTime.Minute(), 0, 0, c.timezone))
	if nextTime.Before(now) {
		nextTime = c.findMatchingDay(nextTime.AddDate(0, 0, 1))
	}

	return time.NewTicker(nextTime.Sub(now))
}

func (c *Cron) hasError(onError func(error)) bool {
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			if onError != nil {
				onError(err)
			}
		}
		return true
	}

	if c.day == 0 && c.interval == 0 {
		if onError != nil {
			onError(errors.New("no schedule configuration"))
		}
		return true
	}

	return false
}

// Start cron
func (c *Cron) Start(action func(time.Time) error, onError func(error)) {
	c.done = make(chan struct{})

	if c.hasError(onError) {
		return
	}

	retryCount := uint(0)
	shouldRetry := false

	for {
		tick := c.getTicker(shouldRetry)

		select {
		case <-c.done:
			tick.Stop()
			return
		case now := <-tick.C:
			if err := action(now); err != nil {
				if onError != nil {
					onError(err)
				}

				retryCount++
				shouldRetry = retryCount <= c.maxRetry
			} else {
				retryCount = 0
				shouldRetry = false
			}
		}
	}
}

// Stop cron
func (c *Cron) Stop() {
	close(c.done)
}

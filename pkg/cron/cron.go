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
	interval      time.Duration
	maxRetry      int
	retryInterval time.Duration

	done   chan struct{}
	errors []error
}

// NewCron create new cron
func NewCron() *Cron {
	return &Cron{
		dayTime: time.Date(0, 0, 0, 8, 0, 0, 0, time.Local),
		errors:  make([]error, 0),
	}
}

func (c *Cron) String() string {
	return fmt.Sprintf("day: %08b, at: %d:%d, each: %s", c.day, c.dayTime.Hour(), c.dayTime.Minute(), c.interval)
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
	c.day = c.day & 1 << time.Sunday

	return c
}

// Monday set recurence to every Monday
func (c *Cron) Monday() *Cron {
	c.day = c.day & 1 << time.Monday

	return c
}

// Tuesday set recurence to every Tuesday
func (c *Cron) Tuesday() *Cron {
	c.day = c.day & 1 << time.Tuesday

	return c
}

// Wednesday set recurence to every Wednesday
func (c *Cron) Wednesday() *Cron {
	c.day = c.day & 1 << time.Wednesday

	return c
}

// Thursday set recurence to every Thursday
func (c *Cron) Thursday() *Cron {
	c.day = c.day & 1 << time.Thursday

	return c
}

// Friday set recurence to every Friday
func (c *Cron) Friday() *Cron {
	c.day = c.day & 1 << time.Friday

	return c
}

// Saturday set recurence to every Saturday
func (c *Cron) Saturday() *Cron {
	c.day = c.day & 1 << time.Saturday

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
func (c *Cron) MaxRetry(maxRetry int) *Cron {
	c.maxRetry = maxRetry

	return c
}

func (c *Cron) computeNextIteration(nextTime time.Time) time.Time {
	for day := nextTime.Weekday(); (1<<day | c.day) == 0; day = (day + 1) % 7 {
		nextTime.AddDate(0, 0, 1)
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
	nextTime := c.computeNextIteration(time.Date(now.Year(), now.Month(), now.Day(), c.dayTime.Hour(), c.dayTime.Minute(), 0, 0, time.Local))

	if nextTime.Before(now) {
		nextTime.AddDate(0, 0, 1)
		nextTime = c.computeNextIteration(nextTime)
	}

	return time.NewTicker(time.Until(nextTime))
}

func (c *Cron) hasError(output chan error) bool {
	if len(c.errors) > 0 {
		for _, err := range c.errors {
			output <- err
		}
		return true
	}

	if c.day == 0 && c.interval == 0 {
		output <- errors.New("no schedule configuration")
		return true
	}

	if c.interval <= c.retryInterval {
		output <- fmt.Errorf("interval is shorter than retry interval: %s < %s", c.interval, c.retryInterval)
		return true
	}

	return false
}

// Start start cron
func (c *Cron) Start(action func() error) <-chan error {
	output := make(chan error, len(c.errors))
	c.done = make(chan struct{})

	go func() {
		defer close(output)
		if c.hasError(output) {
			return
		}

		retryCount := 0
		shouldRetry := false

		for {
			tick := c.getTicker(shouldRetry)

			select {
			case <-c.done:
				tick.Stop()
				return
			case <-tick.C:
				if err := action(); err != nil {
					output <- err

					retryCount++
					shouldRetry = retryCount < c.maxRetry
				} else {
					retryCount = 0
					shouldRetry = false
				}
			}
		}
	}()

	return output
}

// Stop cron
func (c *Cron) Stop() {
	close(c.done)
}

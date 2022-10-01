package clock

import "time"

type Clock struct {
	now time.Time
}

func New(now time.Time) Clock {
	return Clock{
		now: now,
	}
}

func (c Clock) Now() time.Time {
	if c.now.IsZero() {
		return time.Now()
	}

	return c.now
}

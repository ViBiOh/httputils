package clock

import "time"

// Clock give time
type Clock struct {
	now time.Time
}

// New create a new clock
func New(now time.Time) *Clock {
	return &Clock{
		now: now,
	}
}

// Now return current time
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

package cron

import (
	"context"
	"time"
)

// Redis client
type Redis interface {
	Ping() error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) error
}

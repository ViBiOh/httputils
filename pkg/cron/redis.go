package cron

import (
	"context"
	"time"
)

// Redis client
//go:generate mockgen -destination ../mocks/redis.go -mock_names Redis=Redis -package mocks github.com/ViBiOh/httputils/v4/pkg/cron Redis
type Redis interface {
	Ping() error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) error
}

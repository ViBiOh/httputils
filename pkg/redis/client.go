package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client interface {
	Enabled() bool
	Ping(ctx context.Context) error
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Scan(ctx context.Context, pattern string, output chan<- string, pageSize int64) error
	Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Push(ctx context.Context, key string, value any) error
	Pull(ctx context.Context, key string, handler func(string, error))
	Publish(ctx context.Context, channel string, value any) error
	PublishJSON(ctx context.Context, channel string, value any) error
	Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error)
}

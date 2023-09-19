package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Enabled() bool
	Close()
	FlushAll(context.Context) error
	Ping(ctx context.Context) error
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	StoreMany(ctx context.Context, values map[string]any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Scan(ctx context.Context, pattern string, output chan<- string, pageSize int64) error
	Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (bool, error)
	Expire(ctx context.Context, ttl time.Duration, keys ...string) error
	Push(ctx context.Context, key string, value any) error
	Pull(ctx context.Context, key string, handler func(string, error))
	Publish(ctx context.Context, channel string, value any) error
	PublishJSON(ctx context.Context, channel string, value any) error
	Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error)
	Pipeline() redis.Pipeliner
}

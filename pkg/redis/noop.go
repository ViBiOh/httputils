package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrDisabled = errors.New("redis not enabled")

type Noop struct{}

func (n Noop) Enabled() bool {
	return false
}

func (n Noop) Close() {
	// noop
}

func (n Noop) FlushAll(_ context.Context) error {
	return nil
}

func (n Noop) Ping(_ context.Context) error {
	return nil
}

func (n Noop) Load(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (n Noop) LoadMany(_ context.Context, keys ...string) ([]string, error) {
	return make([]string, len(keys)), nil
}

func (n Noop) Store(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (n Noop) StoreMany(_ context.Context, _ map[string]any, _ time.Duration) error {
	return nil
}

func (n Noop) Delete(_ context.Context, _ ...string) error {
	return nil
}

func (n Noop) DeletePattern(_ context.Context, _ string) error {
	return nil
}

func (n Noop) Scan(_ context.Context, _ string, output chan<- string, _ int64) error {
	close(output)

	return nil
}

func (n Noop) Exclusive(_ context.Context, _ string, _ time.Duration, _ func(context.Context) error) (bool, error) {
	return false, ErrDisabled
}

func (n Noop) Expire(_ context.Context, _ time.Duration, _ ...string) error {
	return nil
}

func (n Noop) Push(_ context.Context, _ string, _ any) error {
	return nil
}

func (n Noop) Pull(_ context.Context, _ string, _ func(string, error)) {
	// noop
}

func (n Noop) Publish(_ context.Context, _ string, _ any) error {
	return nil
}

func (n Noop) PublishJSON(_ context.Context, _ string, _ any) error {
	return nil
}

func (n Noop) Subscribe(_ context.Context, _ string) (<-chan *redis.Message, func(context.Context) error) {
	content := make(chan *redis.Message, 1)
	close(content)

	return content, func(_ context.Context) error { return nil }
}

func (n Noop) Pipeline() redis.Pipeliner {
	return nil
}

package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrDisabled = errors.New("redis not enabled")

type noop struct{}

func (n noop) Enabled() bool {
	return false
}

func (n noop) Ping(_ context.Context) error {
	return nil
}

func (n noop) Load(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (n noop) LoadMany(_ context.Context, keys ...string) ([]string, error) {
	return make([]string, len(keys)), nil
}

func (n noop) Store(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (n noop) Delete(_ context.Context, _ ...string) error {
	return nil
}

func (n noop) DeletePattern(_ context.Context, _ string) error {
	return nil
}

func (n noop) Scan(_ context.Context, _ string, output chan<- string, _ int64) error {
	close(output)

	return nil
}

func (n noop) Exclusive(_ context.Context, _ string, _ time.Duration, _ func(context.Context) error) (bool, error) {
	return false, ErrDisabled
}

func (n noop) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (n noop) Push(_ context.Context, _ string, _ any) error {
	return nil
}

func (n noop) Pull(_ context.Context, _ string, _ func(string, error)) {
}

func (n noop) Publish(_ context.Context, _ string, _ any) error {
	return nil
}

func (n noop) PublishJSON(_ context.Context, _ string, _ any) error {
	return nil
}

func (n noop) Subscribe(_ context.Context, _ string) (<-chan *redis.Message, func(context.Context) error) {
	content := make(chan *redis.Message, 1)
	close(content)

	return content, func(_ context.Context) error { return nil }
}

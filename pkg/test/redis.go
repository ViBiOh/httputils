package test

import (
	"context"
	"flag"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/stretchr/testify/assert"
)

type RedisIntegration struct {
	t      *testing.T
	client redis.Client
}

func NewRedisIntegration(t *testing.T) *RedisIntegration {
	t.Helper()

	return &RedisIntegration{t: t}
}

func (ri *RedisIntegration) Bootstrap(ctx context.Context, name string) {
	ri.connect(ctx, name)
}

func (ri *RedisIntegration) connect(ctx context.Context, name string) {
	fs := flag.NewFlagSet("test-"+name, flag.ExitOnError)

	redisConfig := redis.Flags(fs, "")

	if err := fs.Parse(nil); err != nil {
		ri.t.Fatal(err)
	}

	client, err := redis.New(ctx, redisConfig, nil, nil)
	if err != nil {
		ri.t.Fatal(err)
	}

	ri.client = client
}

func (ri *RedisIntegration) Client() redis.Client {
	return ri.client
}

func (ri *RedisIntegration) Reset() {
	err := ri.client.FlushAll(context.Background())

	assert.NoError(ri.t, err)
}

func (ri *RedisIntegration) Close(ctx context.Context) {
	ri.client.Close(ctx)
}

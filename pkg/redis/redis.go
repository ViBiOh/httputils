package redis

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

const (
	metricsNamespace = "redis"
	defaultPageSize  = 100
)

var ErrNoSubscriber = errors.New("no subscriber for channel")

type App struct {
	redisClient redis.UniversalClient
}

type Config struct {
	address  *string
	username *string
	password *string
	alias    *string
	database *int
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		address:  flags.String(fs, prefix, "redis", "Address", "Redis Address host:port (blank to disable)", "localhost:6379", overrides),
		username: flags.String(fs, prefix, "redis", "Username", "Redis Username, if any", "", overrides),
		password: flags.String(fs, prefix, "redis", "Password", "Redis Password, if any", "", overrides),
		database: flags.Int(fs, prefix, "redis", "Database", "Redis Database", 0, overrides),
		alias:    flags.String(fs, prefix, "redis", "Alias", "Connection alias, for metric", "", overrides),
	}
}

func New(config Config, tracer trace.TracerProvider) (Client, error) {
	address := strings.TrimSpace(*config.address)
	if len(address) == 0 {
		return noop{}, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Username: *config.username,
		Password: *config.password,
		DB:       *config.database,
	})

	if err := redisotel.InstrumentTracing(client, redisotel.WithTracerProvider(tracer)); err != nil {
		return noop{}, fmt.Errorf("tracing: %w", err)
	}

	return App{
		redisClient: client,
	}, nil
}

func (a App) Enabled() bool {
	return true
}

func (a App) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return a.redisClient.Ping(ctx).Err()
}

func (a App) Store(ctx context.Context, key string, value any, duration time.Duration) error {
	return a.redisClient.SetEx(ctx, key, value, duration).Err()
}

func (a App) Load(ctx context.Context, key string) ([]byte, error) {
	content, err := a.redisClient.Get(ctx, key).Bytes()
	if err == nil {
		return content, err
	}

	if err != redis.Nil {
		return nil, fmt.Errorf("exec get: %w", err)
	}

	return nil, nil
}

func (a App) LoadMany(ctx context.Context, keys ...string) ([]string, error) {
	pipeline := a.redisClient.Pipeline()

	commands := make([]*redis.StringCmd, len(keys))

	for index, key := range keys {
		commands[index] = pipeline.Get(ctx, key)
	}

	if _, err := pipeline.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("exec pipelined get: %w", err)
	}

	output := make([]string, len(keys))

	for index, result := range commands {
		if result.Err() == nil {
			output[index] = result.Val()
		}
	}

	return output, nil
}

func (a App) Expire(ctx context.Context, ttl time.Duration, keys ...string) error {
	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Expire(ctx, key, ttl)
	}

	return a.execPipeline(ctx, pipeline)
}

func (a App) Delete(ctx context.Context, keys ...string) (err error) {
	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Del(ctx, key)
	}

	return a.execPipeline(ctx, pipeline)
}

func (a App) DeletePattern(ctx context.Context, pattern string) (err error) {
	scanOutput := make(chan string, runtime.NumCPU())

	done := make(chan struct{})

	go func() {
		defer close(done)

		pipeline := a.redisClient.Pipeline()

		for key := range scanOutput {
			pipeline.Del(ctx, key)
		}

		err = a.execPipeline(ctx, pipeline)
	}()

	if err := a.Scan(ctx, pattern, scanOutput, defaultPageSize); err != nil {
		return fmt.Errorf("exec scan: %w", err)
	}

	<-done

	return
}

func (a App) execPipeline(ctx context.Context, pipeline redis.Pipeliner) error {
	results, err := pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("exec pipeline: %w", err)
	}

	for _, result := range results {
		if err = result.Err(); err != nil {
			return fmt.Errorf("pipeline item `%s`: %w", result.Name(), err)
		}
	}

	return nil
}

func (a App) Scan(ctx context.Context, pattern string, output chan<- string, pageSize int64) error {
	defer close(output)

	var keys []string
	var err error
	var cursor uint64

	for {
		keys, cursor, err = a.redisClient.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return fmt.Errorf("exec scan: %w", err)
		}

		for _, key := range keys {
			output <- key
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (a App) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	if acquired, err = a.redisClient.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		err = fmt.Errorf("exec setnx: %w", err)

		return
	} else if !acquired {
		return
	}

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = action(actionCtx)

	if delErr := a.redisClient.Del(ctx, name).Err(); delErr != nil {
		err = errors.Join(err, delErr)
	}

	return
}

func (a App) Pipeline() redis.Pipeliner {
	return a.redisClient.Pipeline()
}

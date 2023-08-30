package redis

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	metricsNamespace = "redis"
	defaultPageSize  = 100
)

var ErrNoSubscriber = errors.New("no subscriber for channel")

type Service struct {
	client    redis.UniversalClient
	isCluster bool
}

type Config struct {
	Username    string
	Password    string
	Address     []string
	Database    int
	PoolSize    int
	MinIdleConn int
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Address", "Redis Address host:port (blank to disable)").Prefix(prefix).DocPrefix("redis").StringSliceVar(fs, &config.Address, []string{"127.0.0.1:6379"}, overrides)
	flags.New("Username", "Redis Username, if any").Prefix(prefix).DocPrefix("redis").StringVar(fs, &config.Username, "", overrides)
	flags.New("Password", "Redis Password, if any").Prefix(prefix).DocPrefix("redis").StringVar(fs, &config.Password, "", overrides)
	flags.New("Database", "Redis Database").Prefix(prefix).DocPrefix("redis").IntVar(fs, &config.Database, 0, overrides)
	flags.New("PoolSize", "Redis Pool Size (default GOMAXPROCS*10)").Prefix(prefix).DocPrefix("redis").IntVar(fs, &config.PoolSize, 0, overrides)
	flags.New("MinIdleConn", "Redis Minimum Idle Connections").Prefix(prefix).DocPrefix("redis").IntVar(fs, &config.MinIdleConn, 0, overrides)

	return &config
}

func New(config *Config, meter metric.MeterProvider, tracer trace.TracerProvider) (Client, error) {
	if len(config.Address) == 0 {
		return Noop{}, nil
	}

	service := &Service{
		isCluster: len(config.Address) > 1,
		client: redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        config.Address,
			Username:     config.Username,
			Password:     config.Password,
			DB:           config.Database,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConn,
		}),
	}

	if !model.IsNil(tracer) {
		if err := redisotel.InstrumentTracing(service.client, redisotel.WithTracerProvider(tracer)); err != nil {
			defer service.Close()

			return Noop{}, fmt.Errorf("tracing: %w", err)
		}
	}

	if !model.IsNil(meter) {
		if err := redisotel.InstrumentMetrics(service.client, redisotel.WithMeterProvider(meter)); err != nil {
			defer service.Close()

			return Noop{}, fmt.Errorf("meter: %w", err)
		}
	}

	return service, nil
}

func (s *Service) Enabled() bool {
	return true
}

func (s *Service) Close() {
	if err := s.client.Close(); err != nil {
		slog.Error("redis close", "err", err)
	}
}

func (s *Service) FlushAll(ctx context.Context) error {
	return s.client.FlushAll(ctx).Err()
}

func (s *Service) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return s.client.Ping(ctx).Err()
}

func (s *Service) Store(ctx context.Context, key string, value any, duration time.Duration) error {
	return s.client.Set(ctx, key, value, duration).Err()
}

func (s *Service) Load(ctx context.Context, key string) ([]byte, error) {
	content, err := s.client.Get(ctx, key).Bytes()
	if err == nil {
		return content, err
	}

	if err != redis.Nil {
		return nil, fmt.Errorf("exec get: %w", err)
	}

	return nil, nil
}

func (s *Service) LoadMany(ctx context.Context, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	if !s.isCluster {
		return s.mget(ctx, keys...)
	}

	return s.pipelinedGet(ctx, keys...)
}

func (s *Service) mget(ctx context.Context, keys ...string) ([]string, error) {
	results, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("mget: %w", err)
	}

	output := make([]string, len(results))

	for index, result := range results {
		if value, ok := result.(string); ok {
			output[index] = value
		}
	}

	return output, nil
}

func (s *Service) pipelinedGet(ctx context.Context, keys ...string) ([]string, error) {
	pipeline := s.client.Pipeline()

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

func (s *Service) Expire(ctx context.Context, ttl time.Duration, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	pipeline := s.client.Pipeline()

	for _, key := range keys {
		pipeline.Expire(ctx, key, ttl)
	}

	return s.execPipeline(ctx, pipeline)
}

func (s *Service) Delete(ctx context.Context, keys ...string) (err error) {
	if len(keys) == 0 {
		return nil
	}

	pipeline := s.client.Pipeline()

	for _, key := range keys {
		pipeline.Del(ctx, key)
	}

	return s.execPipeline(ctx, pipeline)
}

func (s *Service) DeletePattern(ctx context.Context, pattern string) (err error) {
	scanOutput := make(chan string, runtime.NumCPU())

	done := make(chan struct{})

	go func() {
		defer close(done)

		pipeline := s.client.Pipeline()

		for key := range scanOutput {
			pipeline.Del(ctx, key)
		}

		err = s.execPipeline(ctx, pipeline)
	}()

	if err := s.Scan(ctx, pattern, scanOutput, defaultPageSize); err != nil {
		return fmt.Errorf("exec scan: %w", err)
	}

	<-done

	return
}

func (s *Service) execPipeline(ctx context.Context, pipeline redis.Pipeliner) error {
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

func (s *Service) Scan(ctx context.Context, pattern string, output chan<- string, pageSize int64) error {
	defer close(output)

	var keys []string
	var err error
	var cursor uint64

	for {
		keys, cursor, err = s.client.Scan(ctx, cursor, pattern, pageSize).Result()
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

func (s *Service) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	if acquired, err = s.client.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		err = fmt.Errorf("exec setnx: %w", err)

		return
	} else if !acquired {
		return
	}

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = action(actionCtx)

	if delErr := s.client.Del(ctx, name).Err(); delErr != nil {
		err = errors.Join(err, delErr)
	}

	return
}

func (s *Service) Pipeline() redis.Pipeliner {
	return s.client.Pipeline()
}

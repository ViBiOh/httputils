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
	"github.com/ViBiOh/httputils/v4/pkg/model"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	metricsNamespace = "redis"
	defaultPageSize  = 100
)

var ErrNoSubscriber = errors.New("no subscriber for channel")

type App struct {
	tracer      trace.Tracer
	redisClient redis.UniversalClient
	metric      *prometheus.CounterVec
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
		address:  flags.String(fs, prefix, "redis", "Address", "Redis Address fqdn:port (blank to disable)", "localhost:6379", overrides),
		username: flags.String(fs, prefix, "redis", "Username", "Redis Username, if any", "", overrides),
		password: flags.String(fs, prefix, "redis", "Password", "Redis Password, if any", "", overrides),
		database: flags.Int(fs, prefix, "redis", "Database", "Redis Database", 0, overrides),
		alias:    flags.String(fs, prefix, "redis", "Alias", "Connection alias, for metric", "", overrides),
	}
}

func New(config Config, prometheusRegisterer prometheus.Registerer, tracer trace.Tracer) Client {
	address := strings.TrimSpace(*config.address)
	if len(address) == 0 {
		return noop{}
	}

	return App{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     address,
			Username: *config.username,
			Password: *config.password,
			DB:       *config.database,
		}),
		metric: prom.CounterVec(prometheusRegisterer, metricsNamespace, strings.TrimSpace(*config.alias), "item", "state"),
		tracer: tracer,
	}
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
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store", trace.WithAttributes(attribute.String("key", key)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	err := a.redisClient.SetEx(ctx, key, value, duration).Err()

	if err == nil {
		a.increase("store")
	} else {
		a.increase("error")
	}

	return err
}

func (a App) Load(ctx context.Context, key string) ([]byte, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "load", trace.WithAttributes(attribute.String("key", key)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	content, err := a.redisClient.Get(ctx, key).Bytes()

	if err == nil {
		a.increase("hit")

		return content, nil
	}

	if err != redis.Nil {
		a.increase("error")

		return nil, fmt.Errorf("exec get: %w", err)
	}

	a.increase("miss")

	return nil, nil
}

func (a App) LoadMany(ctx context.Context, keys ...string) ([]string, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "load_many", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	pipeline := a.redisClient.Pipeline()

	commands := make([]*redis.StringCmd, len(keys))

	for index, key := range keys {
		commands[index] = pipeline.Get(ctx, key)
	}

	if _, err := pipeline.Exec(ctx); err != nil && err != redis.Nil {
		a.increase("error")

		return nil, fmt.Errorf("exec pipelined get: %w", err)
	}

	a.increase("load_many")

	output := make([]string, len(keys))

	for index, result := range commands {
		if result.Err() == nil {
			a.increase("hit")

			output[index] = result.Val()
		}
	}

	return output, nil
}

func (a App) Expire(ctx context.Context, ttl time.Duration, keys ...string) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "expire", trace.WithAttributes(attribute.String("ttl", ttl.String())), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Expire(ctx, key, ttl)
	}

	return a.execPipeline(ctx, pipeline)
}

func (a App) Delete(ctx context.Context, keys ...string) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "delete", trace.WithAttributes(attribute.StringSlice("keys", keys)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Del(ctx, key)
	}

	return a.execPipeline(ctx, pipeline)
}

func (a App) DeletePattern(ctx context.Context, pattern string) (err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "delete_pattern", trace.WithAttributes(attribute.String("pattenr", pattern)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

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
		a.increase("error")

		return fmt.Errorf("exec pipeline: %w", err)
	}

	for _, result := range results {
		if err = result.Err(); err != nil {
			a.increase("error")

			return fmt.Errorf("pipeline item `%s`: %w", result.Name(), err)
		}

		a.increase("delete")
	}

	return nil
}

func (a App) Scan(ctx context.Context, pattern string, output chan<- string, pageSize int64) error {
	defer close(output)

	var keys []string
	var err error
	var cursor uint64

	ctx, end := tracer.StartSpan(ctx, a.tracer, "scan", trace.WithAttributes(attribute.String("pattern", pattern)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	for {
		keys, cursor, err = a.redisClient.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			a.increase("error")
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
	ctx, end := tracer.StartSpan(ctx, a.tracer, "exclusive", trace.WithAttributes(attribute.String("name", name)), trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	a.increase("exclusive")

	if acquired, err = a.redisClient.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		a.increase("error")
		err = fmt.Errorf("exec setnx: %w", err)

		return
	} else if !acquired {
		return
	}

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = action(actionCtx)

	if delErr := a.redisClient.Del(ctx, name).Err(); delErr != nil {
		a.increase("error")
		err = model.WrapError(err, delErr)
	}

	return
}

func (a App) increase(name string) {
	if a.metric == nil {
		return
	}

	a.metric.WithLabelValues(name).Inc()
}

package redis

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	metricsNamespace = "redis"
)

var ErrNoSubscriber = errors.New("no subscriber for channel")

type App struct {
	tracer      trace.Tracer
	redisClient *redis.Client
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

func New(config Config, prometheusRegisterer prometheus.Registerer, tracer trace.Tracer) App {
	address := strings.TrimSpace(*config.address)
	if len(address) == 0 {
		return App{}
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

func (a App) enabled() bool {
	return a.redisClient != nil
}

func (a App) Ping() error {
	if !a.enabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return a.redisClient.Ping(ctx).Err()
}

func (a App) Store(ctx context.Context, key string, value any, duration time.Duration) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "store", trace.WithAttributes(attribute.String("key", key)))
	defer end()

	err := a.redisClient.SetEX(ctx, key, value, duration).Err()

	if err == nil {
		a.increase("store")
	} else {
		a.increase("error")
	}

	return err
}

func (a App) Load(ctx context.Context, key string) ([]byte, error) {
	if !a.enabled() {
		return nil, nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "load", trace.WithAttributes(attribute.String("key", key)))
	defer end()

	content, err := a.redisClient.Get(ctx, key).Bytes()

	if err == nil {
		a.increase("load")

		return content, nil
	}

	if err != redis.Nil {
		a.increase("error")

		return nil, fmt.Errorf("load: %w", err)
	}

	a.increase("miss")

	return nil, nil
}

func (a App) LoadMany(ctx context.Context, keys ...string) ([]string, error) {
	if !a.enabled() {
		return nil, nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "load_many")
	defer end()

	content, err := a.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		a.increase("error")

		return nil, fmt.Errorf("batch load: %w", err)
	}

	a.increase("load_many")
	output := make([]string, len(content))

	for index, raw := range content {
		value, _ := raw.(string)
		output[index] = value
	}

	return output, nil
}

func (a App) Delete(ctx context.Context, keys ...string) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "delete", trace.WithAttributes(attribute.StringSlice("keys", keys)))
	defer end()

	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Del(ctx, key)
	}

	results, err := pipeline.Exec(ctx)
	if err != nil {
		a.increase("error")

		return fmt.Errorf("exec delete pipeline: %w", err)
	}

	for _, result := range results {
		if err = result.Err(); err != nil {
			a.increase("error")

			return fmt.Errorf("delete key: %w", err)
		}

		a.increase("delete")
	}

	return nil
}

func (a App) Scan(ctx context.Context, pattern string, pageSize int64) (output []string, err error) {
	if !a.enabled() {
		return
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "scan", trace.WithAttributes(attribute.String("pattern", pattern)))
	defer end()

	cursor := uint64(0)

	for {
		var keys []string
		keys, cursor, err = a.redisClient.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			a.increase("error")
			err = fmt.Errorf("exec scan: %w", err)
			return
		}

		output = append(output, keys...)

		if cursor == 0 {
			break
		}
	}

	return
}

func (a App) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	if !a.enabled() {
		return false, fmt.Errorf("redis not enabled")
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "exclusive", trace.WithAttributes(attribute.String("name", name)))
	defer end()

	a.increase("exclusive")

	if acquired, err = a.redisClient.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		a.increase("error")
		err = fmt.Errorf("check semaphore: %w", err)

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

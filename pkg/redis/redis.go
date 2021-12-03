package redis

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace = "redis"
)

// App of package
type App struct {
	redisClient *redis.Client
	metric      *prometheus.CounterVec
}

// Config of package
type Config struct {
	redisAddress  *string
	redisPassword *string
	redisAlias    *string
	redisDatabase *int
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		redisAddress:  flags.New(prefix, "redis", "Address").Default("localhost:6379", overrides).Label("Redis Address").ToString(fs),
		redisPassword: flags.New(prefix, "redis", "Password").Default("", overrides).Label("Redis Password, if any").ToString(fs),
		redisDatabase: flags.New(prefix, "redis", "Database").Default(0, overrides).Label("Redis Database").ToInt(fs),
		redisAlias:    flags.New(prefix, "redis", "Alias").Default("", overrides).Label("Connection alias, for metric").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, prometheusRegisterer prometheus.Registerer) App {
	return App{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     *config.redisAddress,
			Password: *config.redisPassword,
			DB:       *config.redisDatabase,
		}),
		metric: prom.CounterVec(prometheusRegisterer, metricsNamespace, strings.TrimSpace(*config.redisAlias), "item", "state"),
	}
}

// Ping check redis availability
func (a App) Ping() error {
	return a.redisClient.Ping(context.Background()).Err()
}

// Store store give key/val with duration
func (a App) Store(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	err := a.redisClient.SetEX(ctx, key, value, duration).Err()

	if err == nil {
		a.increase("store")
	} else {
		a.increase("error")
	}

	return err
}

// Load given key
func (a App) Load(ctx context.Context, key string) (string, error) {
	content, err := a.redisClient.Get(ctx, key).Result()

	if err == nil {
		a.increase("load")
		return content, nil
	}

	if err != redis.Nil {
		a.increase("error")
		return "", fmt.Errorf("unable to load: %s", err)
	}

	a.increase("miss")
	return "", nil
}

// Delete given key
func (a App) Delete(ctx context.Context, keys ...string) error {
	pipeline := a.redisClient.Pipeline()

	for _, key := range keys {
		pipeline.Del(ctx, key)
	}

	results, err := pipeline.Exec(ctx)
	if err != nil {
		a.increase("error")
		return fmt.Errorf("unable to exec delete pipeline: %s", err)
	}

	for _, result := range results {
		if err = result.Err(); err != nil {
			a.increase("error")
			return fmt.Errorf("unable to delete key: %s", err)
		}

		a.increase("delete")
	}

	return nil
}

// Exclusive get an exclusive lock for given name during duration
func (a App) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	a.increase("exclusive")

	if acquired, err = a.redisClient.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		a.increase("error")
		err = fmt.Errorf("unable to check semaphore: %s", err)
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

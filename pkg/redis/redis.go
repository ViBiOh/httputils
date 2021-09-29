package redis

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

// App of package
type App struct {
	redisClient *redis.Client
	metrics     map[string]prometheus.Counter
}

// Config of package
type Config struct {
	redisAddress  *string
	redisPassword *string
	redisDatabase *int
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		redisAddress:  flags.New(prefix, "redis", "Address").Default("localhost:6379", overrides).Label("Redis Address").ToString(fs),
		redisPassword: flags.New(prefix, "redis", "Password").Default("", overrides).Label("Redis Password, if any").ToString(fs),
		redisDatabase: flags.New(prefix, "redis", "Database").Default(0, overrides).Label("Redis Database").ToInt(fs),
	}
}

// New creates new App from Config
func New(config Config, prometheusRegisterer prometheus.Registerer) (App, error) {
	metrics, err := prom.Counters(prometheusRegisterer, "redis", "", "store", "load", "delete", "exclusive", "error")
	if err != nil {
		return App{}, fmt.Errorf("unable to configure metrics: %s", err)
	}

	return App{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     *config.redisAddress,
			Password: *config.redisPassword,
			DB:       *config.redisDatabase,
		}),
		metrics: metrics,
	}, nil
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
	} else if err != redis.Nil {
		a.increase("error")
	}

	return content, err
}

// Delete given key
func (a App) Delete(ctx context.Context, key string) error {
	err := a.redisClient.Del(ctx, key).Err()

	if err == nil {
		a.increase("delete")
	} else {
		a.increase("error")
	}

	return err
}

// Exclusive get an exclusive lock for given name during duration
func (a App) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) error {
	a.increase("exclusive")

	if acquired, err := a.redisClient.SetNX(ctx, name, "acquired", timeout).Result(); err != nil {
		a.increase("error")
		return fmt.Errorf("unable to check semaphore: %s", err)
	} else if !acquired {
		return nil
	}

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := action(actionCtx)

	if delErr := a.redisClient.Del(ctx, name).Err(); delErr != nil {
		a.increase("error")

		if err == nil {
			err = delErr
		} else {
			err = fmt.Errorf("%s: %w", err, delErr)
		}
	}

	return err
}

func (a App) increase(name string) {
	if gauge, ok := a.metrics[name]; ok {
		gauge.Inc()
	}
}

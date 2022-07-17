package redis

import (
	"context"
	"encoding/json"
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
	"go.opentelemetry.io/otel/trace"
)

const (
	metricsNamespace = "redis"
)

// ErrNoSubscriber occurs when a published message is not received by any subscriber
var ErrNoSubscriber = errors.New("no subscriber for channel")

// App of package
type App struct {
	tracer      trace.Tracer
	redisClient *redis.Client
	metric      *prometheus.CounterVec
}

// Config of package
type Config struct {
	address  *string
	username *string
	password *string
	alias    *string
	database *int
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		address:  flags.String(fs, prefix, "redis", "Address", "Redis Address (blank to disable)", "localhost:6379", overrides),
		username: flags.String(fs, prefix, "redis", "Username", "Redis Username, if any", "", overrides),
		password: flags.String(fs, prefix, "redis", "Password", "Redis Password, if any", "", overrides),
		database: flags.Int(fs, prefix, "redis", "Database", "Redis Database", 0, overrides),
		alias:    flags.String(fs, prefix, "redis", "Alias", "Connection alias, for metric", "", overrides),
	}
}

// New creates new App from Config
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

// Ping check redis availability
func (a App) Ping() error {
	if !a.enabled() {
		return nil
	}

	return a.redisClient.Ping(context.Background()).Err()
}

// Store give key/val with duration
func (a App) Store(ctx context.Context, key string, value any, duration time.Duration) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "store")
	defer end()

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
	if !a.enabled() {
		return "", nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "load")
	defer end()

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

// Delete given keys
func (a App) Delete(ctx context.Context, keys ...string) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "delete")
	defer end()

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

// Publish a message to a given channel
func (a App) Publish(ctx context.Context, channel string, value any) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "publish")
	defer end()

	count, err := a.redisClient.Publish(ctx, channel, value).Result()
	if err != nil {
		a.increase("error")
		return fmt.Errorf("unable to publish: %s", err)
	}

	if count == 0 {
		return ErrNoSubscriber
	}

	return nil
}

// Subscribe to a given channel
func (a App) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error) {
	if !a.enabled() {
		return nil, func(_ context.Context) error { return nil }
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "subscribe")
	defer end()

	pubsub := a.redisClient.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) error {
		return pubsub.Unsubscribe(ctx, channel)
	}
}

// SubscribeFor pubsub with unmarshal of given type
func SubscribeFor[T any](ctx context.Context, app App, channel string, handler func(T, error)) func(context.Context) error {
	subscription, unsubscribe := app.Subscribe(ctx, channel)

	output := make(chan T, len(subscription))

	go func() {
		defer close(output)

		for item := range subscription {
			var instance T
			handler(instance, json.Unmarshal([]byte(item.Payload), &instance))
		}
	}()

	return unsubscribe
}

// Push a task to a list
func (a App) Push(ctx context.Context, key string, value any) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "push")
	defer end()

	if err := a.redisClient.LPush(ctx, key, value); err != nil {
		a.increase("error")
		return fmt.Errorf("unable to push: %s", err)
	}

	return nil
}

// Pull a task from a list
func (a App) Pull(ctx context.Context, key string) (string, error) {
	if !a.enabled() {
		return "", nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "pull")
	defer end()

	content, err := a.redisClient.BRPop(ctx, 0, key).Result()
	if err != nil {
		a.increase("error")
		return "", fmt.Errorf("unable to pull: %s", err)
	}

	if len(content) < 2 {
		return "", nil
	}

	return content[1], err
}

// PullFor pull with unmarshal of given type
func PullFor[T any](ctx context.Context, app App, key string) (output T, err error) {
	var content string

	content, err = app.Pull(ctx, key)
	if err != nil {
		return
	}

	if unmarshalErr := json.Unmarshal([]byte(content), &output); unmarshalErr != nil {
		err = fmt.Errorf("unable to unmarshal: %s", unmarshalErr)
	}

	return
}

// Exclusive get an exclusive lock for given name during duration
func (a App) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	if !a.enabled() {
		return false, fmt.Errorf("redis not enabled")
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "exclusive")
	defer end()

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

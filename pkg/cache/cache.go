package cache

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/uuid"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

// Redis client
type Redis interface {
	Ping() error
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Amqp client
type Amqp interface {
	Ping() error
	Consumer(queueName, topic, exchangeName string, retryDelay time.Duration) (string, error)
	Publisher(exchangeName, exchangeType string, args amqp.Table) error
	Publish(payload amqp.Publishing, exchange string) error
	Listen(queue string) (<-chan amqp.Delivery, error)
}

// App of package
type App struct {
	redisApp Redis
	amqpApp  Amqp
	exchange string
}

// Config of package
type Config struct {
	exchange *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		exchange: flags.New(prefix, "cache", "Exchange").Default("", overrides).Label("Exchange name for distributed cache eviction").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, redisApp Redis, amqpApp Amqp) (App, error) {
	exchangeName := strings.TrimSpace(*config.exchange)

	if redisApp == nil {
		return App{}, errors.New("redis client is required")
	}

	if len(exchangeName) != 0 {
		if amqpApp == nil {
			return App{}, errors.New("amqp client is required")
		}

		if err := amqpApp.Publisher(exchangeName, "fanout", nil); err != nil {
			return App{}, fmt.Errorf("unable to configure cache publisher: %s", err)
		}
	}

	return App{
		redisApp: redisApp,
		amqpApp:  amqpApp,
		exchange: exchangeName,
	}, nil
}

// Enabled checks that requirements are met
func (a App) Enabled() bool {
	return a.redisApp != nil
}

// AmqpEnabled checks that distributed cache eviction is enabled
func (a App) AmqpEnabled() bool {
	return a.amqpApp != nil
}

// ListenEvictions listens on amqp for cache eviction message
func (a App) ListenEvictions(handler func(string)) {
	if !a.AmqpEnabled() {
		logger.Error("no distributed cache eviction configured")
		return
	}

	queueName, err := uuid.New()
	if err != nil {
		logger.Error("unable to generate queue name: %s", err)
		return
	}

	if _, err := a.amqpApp.Consumer(queueName, "", a.exchange, 0); err != nil {
		logger.Error("unable to configure cache consumer: %s", err)
		return
	}

	events, err := a.amqpApp.Listen(queueName)
	if err != nil {
		logger.Error("unable to listen cache eviction on queue `%s`: %s", queueName, err)
		return
	}

	for event := range events {
		handler(string(event.Body))
	}
}

// Evict given key from cache
func (a App) Evict(ctx context.Context, key string) error {
	if err := a.redisApp.Delete(ctx, key); err != nil {
		logger.Error("unable to delete key `%s` from cache: %s", key, err)
	}

	if !a.AmqpEnabled() {
		return nil
	}

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(key),
	}

	if err := a.amqpApp.Publish(message, a.exchange); err != nil {
		return fmt.Errorf("unable to publish eviction message: %s", err)
	}

	return nil
}

// Get an item from cache of by given getter. Refactor with generics.
func (a App) Get(ctx context.Context, key string, getter func() (interface{}, error), newObj func() interface{}) (interface{}, error) {
	content, err := a.redisApp.Load(ctx, key)
	if err == nil {
		obj := newObj()
		if jsonErr := json.Unmarshal([]byte(content), obj); jsonErr != nil {
			logger.Warn("unable to unmarshal content for key `%s`: %s", key, jsonErr)
		} else {
			return obj, nil
		}
	} else if err != redis.Nil {
		logger.Warn("unable to load key `%s` from cache: %s", key, err)
	}

	obj, err := getter()

	if err == nil {
		go func() {
			payload, jsonErr := json.Marshal(obj)
			if jsonErr != nil {
				logger.Warn("unable to marshal content for key `%s`: %s", key, jsonErr)
				return
			}

			if redisErr := a.redisApp.Store(context.Background(), key, payload, 0); redisErr != nil {
				logger.Warn("unable to store key `%s` in cache: %s", key, redisErr)
			}
		}()
	}

	return obj, err
}

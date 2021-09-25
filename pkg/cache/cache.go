package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/uuid"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

// Redis client
type Redis interface {
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Amqp client
type Amqp interface {
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
	mutex    sync.RWMutex
	cache    map[string]interface{}
}

// New creates new App from Config
func New(redisApp Redis, amqpApp Amqp, exchange string) (*App, error) {
	if redisApp == nil && amqpApp == nil {
		return nil, errors.New("redis or amqp is required")
	}

	if len(exchange) != 0 {
		if amqpApp == nil {
			return nil, errors.New("amqp client is required")
		}

		if err := amqpApp.Publisher(exchange, "fanout", nil); err != nil {
			return nil, fmt.Errorf("unable to configure cache publisher: %s", err)
		}
	} else if amqpApp != nil {
		return nil, errors.New("exchange name is required")
	}

	app := App{
		redisApp: redisApp,
		amqpApp:  amqpApp,
		exchange: exchange,
		cache:    make(map[string]interface{}),
	}

	if redisApp == nil {
		go app.listenForEvictions()
	}

	return &app, nil
}

// Enabled checks that requirements are met
func (a *App) Enabled() bool {
	return a.redisApp != nil
}

func (a *App) listenForEvictions() {
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
		a.deleteFromCache(string(event.Body))
	}
}

// Evict given key from cache
func (a *App) Evict(ctx context.Context, key string) error {
	if a.redisApp != nil {
		if err := a.deleteFromRedis(ctx, key); err != nil {
			return err
		}
	} else {
		a.deleteFromCache(key)
	}

	if a.amqpApp == nil {
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

func (a *App) deleteFromCache(key string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.cache, key)
}

func (a *App) deleteFromRedis(ctx context.Context, key string) error {
	if err := a.redisApp.Delete(ctx, key); err != nil {
		return fmt.Errorf("unable to delete key `%s` from cache: %s", key, err)
	}

	return nil
}

// Get an item from cache of by given getter. Refactor with generics.
func (a *App) Get(ctx context.Context, key string, getter func() (interface{}, error), newObj func() interface{}) (interface{}, error) {
	if content := a.get(ctx, key, newObj); content != nil {
		return content, nil
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

func (a *App) get(ctx context.Context, key string, newObj func() interface{}) interface{} {
	if a.redisApp == nil {
		a.mutex.RLock()
		defer a.mutex.RUnlock()
		return a.cache[key]
	}

	content, err := a.redisApp.Load(ctx, key)
	if err == nil {
		obj := newObj()
		if jsonErr := json.Unmarshal([]byte(content), obj); jsonErr != nil {
			logger.Warn("unable to unmarshal content for key `%s`: %s", key, jsonErr)
		} else {
			return obj
		}
	} else if err != redis.Nil {
		logger.Warn("unable to load key `%s` from cache: %s", key, err)
	}

	return nil
}

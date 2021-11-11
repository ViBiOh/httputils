package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// RedisClient for caching response
//go:generate mockgen -destination ../mocks/redis_client.go -mock_names RedisClient=RedisClient -package mocks github.com/ViBiOh/httputils/v4/pkg/cache RedisClient
type RedisClient interface {
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// Retrieve loads an item from the cache for given key or retrieve it (and store it in cache after)
func Retrieve(ctx context.Context, redisClient RedisClient, key string, item interface{}, onMiss func() (interface{}, error), duration time.Duration) (interface{}, error) {
	content, err := redisClient.Load(ctx, key)
	if err != nil {
		logger.Error("unable to read from cache: %s", err)
	} else if len(content) != 0 {
		if err = json.Unmarshal([]byte(content), item); err != nil {
			logger.Error("unable to unmarshal from cache: %s", err)
		} else {
			return item, nil
		}
	}

	item, err = onMiss()

	if err == nil {
		go func() {
			storeCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			if payload, err := json.Marshal(item); err != nil {
				logger.Error("unable to marshal to cache: %s", err)
			} else if err = redisClient.Store(storeCtx, key, payload, duration); err != nil {
				logger.Error("unable to write to cache: %s", err)
			}
		}()
	}

	return item, err
}

// OnModify handle an item update and evict the cache for given key
func OnModify(ctx context.Context, redisClient RedisClient, key string, err error) error {
	if err != nil {
		return err
	}

	if err = redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("unable to evict key `%s` from cache: %s", key, err)
	}

	return nil
}

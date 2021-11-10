package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// RedisClient for caching response
//go:generate mockgen -destination ../mocks/redis_client.go -mock_names RedisClient=RedisClient -package mocks github.com/ViBiOh/httputils/v4/pkg/cache RedisClient
type RedisClient interface {
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value interface{}, duration time.Duration) error
}

// Cacheable object that provide a key
type Cacheable interface {
	GetKey() string
}

// Read an item from the cache for given key or retrieve it (and store it in cache after)
func Read(ctx context.Context, redisClient RedisClient, key string, item Cacheable, onMiss func() (Cacheable, error), duration time.Duration) (Cacheable, error) {
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

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (a App) Push(ctx context.Context, key string, value any) error {
	if content, err := json.Marshal(value); err != nil {
		return fmt.Errorf("marshal: %w", err)
	} else if err := a.redisClient.LPush(ctx, key, content).Err(); err != nil {
		return fmt.Errorf("push: %w", err)
	}

	return nil
}

func (a App) Pull(ctx context.Context, key string, handler func(string, error)) {
	for {
		content, err := a.redisClient.BRPop(ctx, 0, key).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if strings.HasSuffix(err.Error(), "connect: connection refused") {
				time.Sleep(time.Minute)
			} else {
				handler("", err)
			}

			continue
		}

		if len(content) == 2 {
			handler(content[1], nil)
		}
	}
}

func PullFor[T any](ctx context.Context, app Client, key string, handler func(T, error)) {
	app.Pull(ctx, key, func(content string, err error) {
		var instance T

		if err != nil {
			handler(instance, err)
		} else if unmarshalErr := json.Unmarshal([]byte(content), &instance); unmarshalErr != nil {
			handler(instance, fmt.Errorf("unmarshal: %w", unmarshalErr))
		} else {
			handler(instance, nil)
		}
	})
}

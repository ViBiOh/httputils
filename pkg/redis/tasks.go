package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (a Service) Push(ctx context.Context, key string, value any) error {
	if content, err := json.Marshal(value); err != nil {
		return fmt.Errorf("marshal: %w", err)
	} else if err := a.client.LPush(ctx, key, content).Err(); err != nil {
		return fmt.Errorf("push: %w", err)
	}

	return nil
}

func (a Service) Pull(ctx context.Context, key string, handler func(string, error)) {
	for {
		content, err := a.client.BRPop(ctx, 0, key).Result()
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

func PullFor[T any](ctx context.Context, client Client, key string, handler func(T, error)) {
	client.Pull(ctx, key, func(content string, err error) {
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

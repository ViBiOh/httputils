package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
)

// Push a task to a list
func (a App) Push(ctx context.Context, key string, value any) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "push")
	defer end()

	if content, err := json.Marshal(value); err != nil {
		return fmt.Errorf("marshal: %s", err)
	} else if err := a.redisClient.LPush(ctx, key, content).Err(); err != nil {
		a.increase("error")

		return fmt.Errorf("push: %s", err)
	}

	return nil
}

// Pull tasks from a list
func (a App) Pull(ctx context.Context, key string, done <-chan struct{}, handler func(string, error)) {
	if !a.enabled() {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
			cancel()
		}
	}()

	for {
		content, err := a.redisClient.BRPop(ctx, 0, key).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if strings.HasSuffix(err.Error(), "connect: connection refused") {
				time.Sleep(time.Minute)
			} else {
				a.increase("error")
				handler("", err)
			}

			continue
		}

		if len(content) == 2 {
			handler(content[1], nil)
		}
	}
}

// PullFor pull with unmarshal of given type
func PullFor[T any](ctx context.Context, app App, key string, done <-chan struct{}, handler func(T, error)) {
	app.Pull(ctx, key, done, func(content string, err error) {
		var instance T

		if err != nil {
			handler(instance, err)
		} else if unmarshalErr := json.Unmarshal([]byte(content), &instance); unmarshalErr != nil {
			handler(instance, fmt.Errorf("unmarshal: %s", unmarshalErr))
		} else {
			handler(instance, nil)
		}
	})
}

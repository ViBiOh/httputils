package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

func (a App) Push(ctx context.Context, key string, value any) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "push", trace.WithSpanKind(trace.SpanKindProducer))
	defer end()

	if content, err := json.Marshal(value); err != nil {
		return fmt.Errorf("marshal: %w", err)
	} else if err := a.redisClient.LPush(ctx, key, content).Err(); err != nil {
		a.increase("error")

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

func PullFor[T any](ctx context.Context, app App, key string, handler func(T, error)) {
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

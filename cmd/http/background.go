package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	amqp "github.com/rabbitmq/amqp091-go"
)

func startBackground(ctx context.Context, clients clients, adapters adapters) {
	go redis.SubscribeFor(ctx, clients.redis, "httputils:tasks", func(content time.Time, err error) {
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "consume on pubsub", slog.Any("error", err))

			return
		}

		slog.LogAttrs(ctx, slog.LevelInfo, "time from pubsub", slog.Time("content", content))
	})

	speakingClock := cron.New().Each(15 * time.Second).OnSignal(syscall.SIGUSR1).OnError(func(ctx context.Context, err error) {
		slog.LogAttrs(ctx, slog.LevelError, "run cron", slog.Any("error", err))
	}).Now()

	go speakingClock.Start(ctx, func(_ context.Context) error {
		slog.InfoContext(ctx, "Clock is ticking")

		if err := clients.redis.PublishJSON(ctx, "httputils:tasks", time.Now()); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "publish on pubsub", slog.Any("error", err))
		}

		return nil
	})

	go adapters.amqp.Start(ctx)
}

func amqpHandler(_ context.Context, message amqp.Delivery) error {
	var payload map[string]any
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	return nil
}

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

func startBackground(ctx context.Context, config configuration, client client, adapter adapter) func() {
	var closers []func()

	go redis.SubscribeFor(ctx, client.redis, "httputils:tasks", func(content time.Time, err error) {
		if err != nil {
			slog.Error("consume on pubsub", "err", err)

			return
		}

		slog.Info("time from pubsub", "content", content)
	})

	speakingClock := cron.New().Each(15 * time.Second).OnSignal(syscall.SIGUSR1).OnError(func(err error) {
		slog.Error("run cron", "err", err)
	}).Now()

	go speakingClock.Start(ctx, func(_ context.Context) error {
		slog.Info("Clock is ticking")

		if err := client.redis.PublishJSON(ctx, "httputils:tasks", time.Now()); err != nil {
			slog.Error("publish on pubsub", "err", err)
		}

		return nil
	})

	closers = append(closers, speakingClock.Shutdown)

	go adapter.amqp.Start(ctx)

	return func() {
		for _, closer := range closers {
			closer()
		}
	}
}

func amqpHandler(_ context.Context, message amqp.Delivery) error {
	var payload map[string]any
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	return nil
}

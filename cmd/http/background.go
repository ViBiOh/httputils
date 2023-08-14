package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	amqp "github.com/rabbitmq/amqp091-go"
)

func startBackground(ctx context.Context, config configuration, client client, adapter adapter) func() {
	ctx = client.health.Done(ctx)

	var closers []func()

	go client.redis.Pull(ctx, "httputils:tasks", func(content string, err error) {
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info("content=`" + content + "`")
	})

	speakingClock := cron.New().Each(5 * time.Minute).OnSignal(syscall.SIGUSR1).OnError(func(err error) {
		slog.Error("run cron", "err", err)
	}).Now()

	go speakingClock.Start(ctx, func(_ context.Context) error {
		slog.Info("Clock is ticking")

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

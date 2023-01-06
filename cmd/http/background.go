package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	amqplib "github.com/streadway/amqp"
)

func startBackground(ctx context.Context, config configuration, client client, adapter adapter) func() {
	ctx = client.health.Done(ctx)

	var closers []func()

	go client.redis.Pull(ctx, "httputils:tasks", func(content string, err error) {
		if err != nil {
			logger.Fatal(err)
		}

		logger.Info("content=`%s`", content)
	})

	speakingClock := cron.New().Each(5 * time.Minute).OnSignal(syscall.SIGUSR1).OnError(func(err error) {
		logger.Error("error while running cron: %s", err)
	}).Now()

	go speakingClock.Start(ctx, func(_ context.Context) error {
		logger.Info("Clock is ticking")

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

func amqpHandler(_ context.Context, message amqplib.Delivery) error {
	var payload map[string]any
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	return nil
}

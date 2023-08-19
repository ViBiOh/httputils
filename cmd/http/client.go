package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

type client struct {
	redis     redis.Client
	telemetry telemetry.App
	amqp      *amqp.Client
	health    *health.App
}

const closeTimeout = time.Second * 10

func newClient(ctx context.Context, config configuration) (client, error) {
	var output client
	var err error

	logger.Init(config.logger)

	output.telemetry, err = telemetry.New(ctx, config.telemetry)
	if err != nil {
		return output, fmt.Errorf("telemetry: %w", err)
	}

	request.AddOpenTelemetryToDefaultClient(output.telemetry.MeterProvider(), output.telemetry.TracerProvider())

	output.health = health.New(config.health)

	output.redis, err = redis.New(config.redis, output.telemetry.MeterProvider(), output.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("redis: %w", err)
	}

	output.amqp, err = amqp.New(config.amqp, output.telemetry.MeterProvider(), output.telemetry.TracerProvider())
	if err != nil && !errors.Is(err, amqp.ErrNoConfig) {
		return output, fmt.Errorf("amqp: %w", err)
	}

	return output, nil
}

func (c client) Close(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, closeTimeout)
	defer cancel()

	c.amqp.Close()
	c.redis.Close()
	c.telemetry.Close(ctx)
}

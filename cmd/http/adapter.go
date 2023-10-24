package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

type adapter struct {
	renderer *renderer.Service
	amqp     *amqphandler.Service
	hello    *cache.Cache[string, string]
}

func newAdapter(ctx context.Context, config configuration, client client) (adapter, error) {
	var output adapter
	var err error

	output.amqp, err = amqphandler.New(config.amqHandler, client.amqp, client.telemetry.MeterProvider(), client.telemetry.TracerProvider(), amqpHandler)
	if err != nil {
		return output, fmt.Errorf("amqphandler: %w", err)
	}

	output.renderer, err = renderer.New(config.renderer, content, nil, client.telemetry.MeterProvider(), client.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	output.hello = cache.New(client.redis, func(id string) string { return id }, func(ctx context.Context, id string) (string, error) { return hash.String(id), nil }, client.telemetry.TracerProvider()).
		WithTTL(time.Hour).
		WithExtendOnHit(ctx, 10*time.Second).
		WithClientSideCaching(ctx, "httputils_hello", 10)

	return output, nil
}

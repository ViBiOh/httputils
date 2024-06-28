package main

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

//go:embed templates static
var content embed.FS

type adapters struct {
	amqp     *amqphandler.Service
	renderer *renderer.Service
	hello    *cache.Cache[string, string]
}

func newAdapters(ctx context.Context, config configuration, clients clients) (adapters, error) {
	var output adapters
	var err error

	if clients.amqp != nil {
		if err = clients.amqp.Publisher(config.amqHandler.Exchange, "direct", nil); err != nil {
			return output, fmt.Errorf("publisher: %w", err)
		}
	}

	output.amqp, err = amqphandler.New(config.amqHandler, clients.amqp, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider(), amqpHandler)
	if err != nil {
		return output, fmt.Errorf("amqphandler: %w", err)
	}

	output.renderer, err = renderer.New(ctx, config.renderer, content, nil, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	output.hello = cache.New(clients.redis, func(id string) string { return id }, func(_ context.Context, id string) (string, error) { return hash.String(id), nil }, clients.telemetry.TracerProvider()).
		WithTTL(time.Hour).
		WithExtendOnHit(ctx, 10*time.Second, 50).
		WithClientSideCaching(ctx, "httputils_hello", 10)

	return output, nil
}

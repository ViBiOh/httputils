package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

//go:embed templates static
var content embed.FS

func main() {
	config, err := newConfig()
	if err != nil {
		logger.Fatal(fmt.Errorf("config: %w", err))
	}

	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	client, err := newClient(ctx, config)
	if err != nil {
		logger.Fatal(fmt.Errorf("client: %w", err))
	}

	defer client.Close(ctx)

	adapter, err := newAdapter(config, client)
	if err != nil {
		logger.Fatal(fmt.Errorf("adapter: %w", err))
	}

	defer newBackground(config, client, adapter)
	handler := newPort(config, client, adapter)

	appServer := server.New(config.appServer)
	promServer := server.New(config.promServer)

	ctxEnd := client.health.ContextEnd()

	go promServer.Start(ctxEnd, "prometheus", client.prometheus.Handler())
	go appServer.Start(ctxEnd, "http", httputils.Handler(adapter.renderer.Handler(handler.template), client.health, recoverer.Middleware, client.prometheus.Middleware, client.tracer.Middleware, owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), promServer.Done(), adapter.amqp.Done())
}

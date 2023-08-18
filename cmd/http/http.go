package main

import (
	"context"
	"embed"
	"log/slog"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

//go:embed templates static
var content embed.FS

func main() {
	config, err := newConfig()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	client, err := newClient(ctx, config)
	if err != nil {
		slog.Error("client", "err", err)
		os.Exit(1)
	}

	defer client.Close(ctx)

	adapter, err := newAdapter(config, client)
	if err != nil {
		slog.Error("adapter", "err", err)
		os.Exit(1)
	}

	stopBackground := startBackground(ctx, config, client, adapter)
	defer stopBackground()

	handler := newPort(config, client, adapter)

	appServer := server.New(config.appServer)

	ctxEnd := client.health.End(ctx)

	go appServer.Start(ctxEnd, "http", httputils.Handler(adapter.renderer.Handler(handler.template), client.health, recoverer.Middleware, client.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), adapter.amqp.Done())
}

package main

import (
	"context"
	"embed"
	"syscall"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

//go:embed templates static
var content embed.FS

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	client, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")

	go client.Start()
	defer client.Close(ctx)

	ctxEnd := client.health.EndCtx()

	adapter, err := newAdapter(ctxEnd, config, client)
	logger.FatalfOnErr(ctx, err, "adapter")

	startBackground(ctxEnd, client, adapter)

	handler := newPort(config, client, adapter)

	appServer := server.New(config.appServer)

	go appServer.Start(ctxEnd, httputils.Handler(adapter.renderer.Handler(handler.template), client.health, client.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done(), syscall.SIGTERM, syscall.SIGINT)
	server.GracefulWait(appServer.Done(), adapter.amqp.Done())
}

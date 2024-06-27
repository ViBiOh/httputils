package main

import (
	"context"
	"syscall"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")

	go clients.Start()
	defer clients.Close(ctx)

	ctxEnd := clients.health.EndCtx()

	adapters, err := newAdapters(ctxEnd, config, clients)
	logger.FatalfOnErr(ctx, err, "adapter")

	startBackground(ctxEnd, clients, adapters)

	services := newServices(config)
	port := newPort(config, clients, adapters)

	go services.server.Start(ctxEnd, httputils.Handler(port, clients.health, clients.telemetry.Middleware("http"), services.owasp.Middleware, services.cors.Middleware))

	clients.health.WaitForTermination(services.server.Done(), syscall.SIGTERM, syscall.SIGINT)
	server.GracefulWait(services.server.Done(), adapters.amqp.Done())
}

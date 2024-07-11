package main

import (
	"context"
	"syscall"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")

	go clients.Start()
	defer clients.Close(ctx)

	adapters, err := newAdapters(clients.health.EndCtx(), config, clients)
	logger.FatalfOnErr(ctx, err, "adapter")

	startBackground(clients.health.EndCtx(), clients, adapters)

	services := newServices(config)
	port := newPort(config, clients, adapters, services)

	go services.server.Start(clients.health.EndCtx(), port)

	clients.health.WaitForTermination(services.server.Done(), syscall.SIGTERM, syscall.SIGINT)
	health.WaitAll(services.server.Done(), adapters.amqp.Done())
}

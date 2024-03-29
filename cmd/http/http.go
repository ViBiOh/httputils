package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"syscall"

	_ "net/http/pprof"

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
	ctx := context.Background()

	config, err := newConfig()
	logger.FatalfOnErr(ctx, err, "config")

	alcotest.DoAndExit(config.alcotest)

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	client, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")

	defer client.Close(ctx)

	ctxEnd := client.health.EndCtx()

	adapter, err := newAdapter(ctxEnd, config, client)
	logger.FatalfOnErr(ctx, err, "adapter")

	startBackground(ctxEnd, config, client, adapter)

	handler := newPort(ctxEnd, config, client, adapter)

	appServer := server.New(config.appServer)

	go appServer.Start(ctxEnd, httputils.Handler(adapter.renderer.Handler(handler.template), client.health, recoverer.Middleware, client.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done(), syscall.SIGTERM, syscall.SIGINT)
	server.GracefulWait(appServer.Done(), adapter.amqp.Done())
}

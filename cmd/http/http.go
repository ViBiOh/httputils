package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"

	_ "net/http/pprof"

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
	ctx := context.Background()

	config, err := newConfig()
	if err != nil {
		slog.ErrorContext(ctx, "config", "err", err)
		os.Exit(1)
	}

	alcotest.DoAndExit(config.alcotest)

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	client, err := newClient(ctx, config)
	if err != nil {
		slog.ErrorContext(ctx, "client", "err", err)
		os.Exit(1)
	}

	defer client.Close(ctx)

	ctxEnd := client.health.EndCtx()

	adapter, err := newAdapter(ctxEnd, config, client)
	if err != nil {
		slog.ErrorContext(ctx, "adapter", "err", err)
		os.Exit(1)
	}

	startBackground(ctxEnd, config, client, adapter)

	handler := newPort(ctxEnd, config, client, adapter)

	appServer := server.New(config.appServer)

	go appServer.Start(ctxEnd, httputils.Handler(adapter.renderer.Handler(handler.template), client.health, recoverer.Middleware, client.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done(), syscall.SIGTERM, syscall.SIGINT)
	server.GracefulWait(appServer.Done(), adapter.amqp.Done())
}

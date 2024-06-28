package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func newPort(config configuration, clients clients, adapters adapters, services services) http.Handler {
	mux := http.NewServeMux()

	adapters.renderer.RegisterMux(mux, getDefaultRenderer(config, clients, adapters))

	return httputils.Handler(mux, clients.health,
		clients.telemetry.Middleware("http"),
		services.owasp.Middleware,
		services.cors.Middleware,
	)
}

func getDefaultRenderer(config configuration, clients clients, adapters adapters) func(http.ResponseWriter, *http.Request) (renderer.Page, error) {
	portTracer := clients.telemetry.TracerProvider().Tracer("port")

	return func(_ http.ResponseWriter, r *http.Request) (renderer.Page, error) {
		var err error

		ctx, end := telemetry.StartSpan(r.Context(), portTracer, "handler", trace.WithSpanKind(trace.SpanKindInternal))
		defer end(&err)

		resp, err := request.Get("https://api.vibioh.fr/dump/").Send(ctx, nil)
		if err != nil {
			return renderer.Page{}, err
		}

		if err = request.DiscardBody(resp.Body); err != nil {
			return renderer.Page{}, err
		}

		if _, err = adapters.hello.Get(ctx, hash.String(r.URL.Path)); err != nil {
			return renderer.Page{}, err
		}

		if err := clients.amqp.PublishJSON(ctx, r.URL.Path, config.amqHandler.Exchange, ""); err != nil {
			return renderer.Page{}, fmt.Errorf("amqp publish: %w", err)
		}

		if len(r.URL.Query().Get("evict")) > 0 {
			go func() {
				time.Sleep(time.Millisecond * 100)
				if err = adapters.hello.EvictOnSuccess(context.WithoutCancel(ctx), r.URL.Path, nil); err != nil {
					slog.LogAttrs(ctx, slog.LevelError, "evict on success", slog.Any("error", err))
				}
			}()
		}

		telemetry.SetRouteTag(ctx, "hello")
		slog.InfoContext(ctx, "Hello World")

		return renderer.NewPage("public", http.StatusOK, nil), nil
	}
}

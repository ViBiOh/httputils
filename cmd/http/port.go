package main

import (
	"context"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

type port struct {
	template renderer.TemplateFunc
}

func newPort(config configuration, client client, adapter adapter) port {
	var output port

	cacheTracer := client.tracer.GetTracer("port")

	simpleCache := cache.New(client.redis, func(id string) string { return id }, func(ctx context.Context, id string) (string, error) {
		_, end := tracer.StartSpan(ctx, cacheTracer, "onMiss", trace.WithSpanKind(trace.SpanKindInternal))
		defer end()

		return id, nil
	}, time.Hour, 4, cacheTracer)

	output.template = func(w http.ResponseWriter, r *http.Request) (renderer.Page, error) {
		resp, err := request.Get("https://api.vibioh.fr/dump/").Send(r.Context(), nil)
		if err != nil {
			return renderer.Page{}, err
		}

		if err = request.DiscardBody(resp.Body); err != nil {
			return renderer.Page{}, err
		}

		if _, err = simpleCache.Get(r.Context(), r.URL.Path); err != nil {
			return renderer.Page{}, err
		}

		return renderer.NewPage("public", http.StatusOK, nil), nil
	}

	return output
}

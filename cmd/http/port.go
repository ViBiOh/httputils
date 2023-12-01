package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type port struct {
	template renderer.TemplateFunc
}

func newPort(ctx context.Context, config configuration, client client, adapter adapter) port {
	var output port

	portTracer := client.telemetry.TracerProvider().Tracer("port")

	output.template = func(w http.ResponseWriter, r *http.Request) (renderer.Page, error) {
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

		if _, err = adapter.hello.Get(r.Context(), hash.String(r.URL.Path)); err != nil {
			return renderer.Page{}, err
		}

		if len(r.URL.Query().Get("evict")) > 0 {
			go func() {
				time.Sleep(time.Millisecond * 100)
				if err = adapter.hello.EvictOnSuccess(cntxt.WithoutDeadline(ctx), r.URL.Path, nil); err != nil {
					slog.ErrorContext(r.Context(), "evict on success", "err", err)
				}
			}()
		}

		return renderer.NewPage("public", http.StatusOK, nil), nil
	}

	return output
}

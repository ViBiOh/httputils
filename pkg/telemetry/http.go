package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type asyncRouteKey struct{}

type tagSetter func(string)

func AsyncRouteTagMiddleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var setter tagSetter = func(route string) {
			attr := semconv.HTTPRouteKey.String(route)

			span := trace.SpanFromContext(ctx)
			span.SetAttributes(attr)

			labeler, _ := otelhttp.LabelerFromContext(ctx)
			labeler.Add(attr)
		}

		newCtx := context.WithValue(ctx, asyncRouteKey{}, setter)

		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func SetRouteTag(ctx context.Context, route string) {
	value := ctx.Value(asyncRouteKey{})
	if value == nil {
		return
	}

	setter, ok := value.(tagSetter)
	if !ok {
		return
	}

	setter(route)
}

package opentracing

import (
	"context"

	"github.com/ViBiOh/httputils/pkg/errors"
	opentracing "github.com/opentracing/opentracing-go"
)

// InjectSpanToMap extract span from map
func InjectSpanToMap(ctx context.Context, content map[string]string) error {
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return nil
	}

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	err := tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(content))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// ExtractSpanFromMap extract span from map
func ExtractSpanFromMap(ctx context.Context, content map[string]string, name string) (context.Context, opentracing.Span, error) {
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return ctx, nil, nil
	}

	spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(content))
	if err != nil {
		return ctx, nil, nil
	}

	span := opentracing.StartSpan(name, opentracing.ChildOf(spanCtx))
	return opentracing.ContextWithSpan(ctx, span), span, nil
}

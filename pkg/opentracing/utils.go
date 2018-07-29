package opentracing

import (
	"context"
	"fmt"

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
		return fmt.Errorf(`Error while injecting span to map: %v`, err)
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
		return ctx, nil, fmt.Errorf(`Error while extracting span from map: %v - %+v`, err, content)
	}

	span := opentracing.StartSpan(name, opentracing.ChildOf(spanCtx))
	return opentracing.ContextWithSpan(ctx, span), span, nil
}

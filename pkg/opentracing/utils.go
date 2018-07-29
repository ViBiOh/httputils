package opentracing

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
)

// InjectSpanToMap extract span from map
func InjectSpanToMap(tracer opentracing.Tracer, spanCtx opentracing.SpanContext, content map[string]string) error {
	err := tracer.Inject(spanCtx, opentracing.TextMap, opentracing.TextMapCarrier(content))
	if err != nil {
		return fmt.Errorf(`Error while injecting span to map: %v`, err)
	}

	return nil
}

// ExtractSpanFromMap extract span from map
func ExtractSpanFromMap(ctx context.Context, tracer opentracing.Tracer, content map[string]string, name string) (context.Context, error) {
	spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(content))
	if err != nil {
		return ctx, fmt.Errorf(`Error while extracting span from map: %v - %+v`, err, content)
	}

	span := opentracing.StartSpan(name, opentracing.ChildOf(spanCtx))
	return opentracing.ContextWithSpan(ctx, span), nil
}

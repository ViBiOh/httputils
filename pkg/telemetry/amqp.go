package telemetry

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleHeaderCarrier map[string]any

func (shc SimpleHeaderCarrier) Get(key string) string {
	raw, ok := shc[key]
	if !ok {
		return ""
	}

	value, ok := raw.(string)
	if !ok {
		return ""
	}

	return value
}

func (shc SimpleHeaderCarrier) Set(key string, value string) {
	shc[key] = value
}

func (shc SimpleHeaderCarrier) Keys() []string {
	keys := make([]string, len(shc))

	index := 0
	for k := range shc {
		keys[index] = k
		index++
	}

	return keys
}

func InjectToAmqp(ctx context.Context, payload amqp.Publishing) amqp.Publishing {
	headers := SimpleHeaderCarrier{}

	propagator.Inject(ctx, headers)

	if model.IsNil(payload.Headers) {
		payload.Headers = amqp.Table{}
	}

	for key, value := range headers {
		payload.Headers[key] = value
	}

	return payload
}

func ExtractContext(ctx context.Context, headers map[string]any) context.Context {
	return propagator.Extract(ctx, SimpleHeaderCarrier(headers))
}

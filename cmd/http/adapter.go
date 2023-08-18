package main

import (
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

type adapter struct {
	renderer *renderer.App
	amqp     *amqphandler.App
}

func newAdapter(config configuration, client client) (adapter, error) {
	var output adapter
	var err error

	output.amqp, err = amqphandler.New(config.amqHandler, client.amqp, client.telemetry.GetMeterProvider(), client.telemetry.GetTracer("amqp_handler"), amqpHandler)
	if err != nil {
		return output, fmt.Errorf("amqphandler: %w", err)
	}

	output.renderer, err = renderer.New(config.renderer, content, nil, client.telemetry.GetMeterProvider(), client.telemetry.GetTracer("renderer"))
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	return output, nil
}

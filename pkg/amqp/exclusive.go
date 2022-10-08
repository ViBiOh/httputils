package amqp

import (
	"context"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/trace"
)

func (c *Client) SetupExclusive(name string) (err error) {
	create, count := c.shouldCreateExclusiveQueue(name)
	if !create && count > 0 {
		return nil
	}

	channel, err := c.connection.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	if create {
		if _, err = channel.QueueDeclare(name, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue: %w", err)
		}
	}

	if err = channel.Publish("", name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("semaphore"),
	}); err != nil {
		return fmt.Errorf("publish semaphore: %w", err)
	}

	return nil
}

func (c *Client) shouldCreateExclusiveQueue(name string) (bool, int) {
	channel, err := c.connection.Channel()
	if err != nil {
		return false, 0
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	queue, err := channel.QueueInspect(name)
	if err != nil {
		return true, 0
	}

	return false, queue.Messages
}

func (c *Client) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	ctx, end := tracer.StartSpan(ctx, c.tracer, "exclusive", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	var message amqp.Delivery
	if message, acquired, err = channel.Get(name, false); err != nil {
		err = fmt.Errorf("get semaphore: %w", err)

		return
	} else if !acquired {
		return
	}

	defer func() {
		if nackErr := message.Nack(false, true); nackErr != nil {
			err = model.WrapError(err, fmt.Errorf("nack message: %w", err))
		}
	}()

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = action(actionCtx)

	return
}

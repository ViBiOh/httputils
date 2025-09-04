package amqp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

func (c *Client) SetupExclusive(ctx context.Context, name string) (err error) {
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

	if err = channel.PublishWithContext(ctx, "", name, false, false, amqp.Publishing{
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

	queue, err := channel.QueueDeclarePassive(name, true, false, false, false, nil)
	if err != nil {
		return true, 0
	}

	return false, queue.Messages
}

func (c *Client) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	ctx, end := telemetry.StartSpan(ctx, c.tracer, "receive", trace.WithSpanKind(trace.SpanKindConsumer))
	defer end(&err)

	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return acquired, err
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	var message amqp.Delivery
	if message, acquired, err = channel.Get(name, false); err != nil {
		return acquired, fmt.Errorf("get semaphore: %w", err)
	} else if !acquired {
		return acquired, err
	}

	defer func() {
		if nackErr := message.Nack(false, true); nackErr != nil {
			err = errors.Join(err, fmt.Errorf("nack message: %w", err))
		}
	}()

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return acquired, action(actionCtx)
}

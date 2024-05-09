package amqp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const reconnectInterval = time.Second * 30

type QueueResolver func() (string, error)

func (c *Client) Listen(queueResolver QueueResolver, exchange, routingKey string) (string, <-chan amqp.Delivery, error) {
	queueName, err := queueResolver()
	if err != nil {
		return "", nil, fmt.Errorf("get queue name: %w", err)
	}

	ctx := context.Background()

	listener, err := c.getListener(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("get listener name for queue `%s`: %w", queueName, err)
	}

	messages, err := c.listen(listener, queueName)
	if err != nil {
		return "", nil, err
	}

	forward := make(chan amqp.Delivery)
	go c.forward(ctx, listener, queueResolver, messages, forward, exchange, routingKey)

	return listener.name, forward, nil
}

func (c *Client) StopListener(consumer string) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	listener := c.listeners[consumer]
	if listener == nil {
		return nil
	}

	if cancelErr := listener.cancel(); cancelErr != nil {
		err = fmt.Errorf("cancel listener: %w", err)
	}

	if closeErr := listener.close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close listener: %w", closeErr))
	}

	c.removeListener(context.Background(), consumer)

	return err
}

func (c *Client) listen(listener *listener, queue string) (<-chan amqp.Delivery, error) {
	if listener.channel == nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()

		if err := listener.createChannel(c.connection); err != nil {
			return nil, err
		}
	}

	listener.RLock()
	defer listener.RUnlock()

	messages, err := listener.channel.Consume(queue, listener.name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume queue: %w", err)
	}

	return messages, nil
}

func (c *Client) forward(ctx context.Context, listener *listener, queueResolver QueueResolver, input <-chan amqp.Delivery, output chan<- amqp.Delivery, exchange, routingKey string) {
	defer close(listener.done)
	defer close(output)

	attributes := append([]attribute.KeyValue{
		semconv.MessagingOperationReceive,
	}, c.getAttributes(exchange, routingKey)...)

forward:
	for delivery := range input {
		c.increase(ctx, attributes)
		output <- delivery
	}

	if _, ok := <-listener.reconnect; !ok {
		return
	}

reconnect:
	log := slog.With("name", listener.name)

	if queueName, err := queueResolver(); err != nil {
		log.LogAttrs(ctx, slog.LevelError, "get queue name on reopen", slog.Any("error", err))
	} else if messages, err := c.listen(listener, queueName); err != nil {
		log.LogAttrs(ctx, slog.LevelError, "reopen listener", slog.Any("error", err))
	} else {
		log.LogAttrs(ctx, slog.LevelInfo, "Listen restarted")
		input = messages

		goto forward
	}

	log.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Waiting %s before attempting to listen again...", reconnectInterval))
	time.Sleep(reconnectInterval)

	goto reconnect
}

package amqp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueResolver func() (string, error)

func (c *Client) Listen(queueResolver QueueResolver, exchange, routingKey string) (string, <-chan amqp.Delivery, error) {
	queueName, err := queueResolver()
	if err != nil {
		return "", nil, fmt.Errorf("get queue name: %w", err)
	}

	listener, err := c.getListener()
	if err != nil {
		return "", nil, fmt.Errorf("get listener name for queue `%s`: %w", queueName, err)
	}

	messages, err := c.listen(listener, queueName)
	if err != nil {
		return "", nil, err
	}

	forward := make(chan amqp.Delivery)
	go c.forward(listener, queueResolver, messages, forward, exchange, routingKey)

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

	c.removeListener(consumer)

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

func (c *Client) forward(listener *listener, queueResolver QueueResolver, input <-chan amqp.Delivery, output chan<- amqp.Delivery, exchange, routingKey string) {
	defer close(listener.done)
	defer close(output)

forward:
	for delivery := range input {
		c.increase(context.Background(), "consume", exchange, routingKey)
		output <- delivery
	}

	if _, ok := <-listener.reconnect; !ok {
		return
	}

reconnect:
	log := slog.With("name", listener.name)

	if queueName, err := queueResolver(); err != nil {
		log.Error("get queue name on reopen", "err", err)
	} else if messages, err := c.listen(listener, queueName); err != nil {
		log.Error("reopen listener", "err", err)
	} else {
		log.Info("Listen restarted")
		input = messages

		goto forward
	}

	log.Info("Waiting 30 seconds before attempting to listen again...")
	time.Sleep(time.Second * 30)

	goto reconnect
}

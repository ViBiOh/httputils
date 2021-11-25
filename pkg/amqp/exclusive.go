package amqp

import (
	"context"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

// SetupExclusive configure the exclusive queue
func (c *Client) SetupExclusive(name string) (err error) {
	create, count := c.shouldCreateExclusiveQueue(name)
	if !create || count > 0 {
		return nil
	}

	channel, err := c.connection.Channel()
	if err != nil {
		return fmt.Errorf("unable to open channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}
	}()

	if create {
		if _, err = channel.QueueDeclare(name, true, false, false, false, nil); err != nil {
			return fmt.Errorf("unable to declare queue: %s", err)
		}
	}

	if err = channel.Publish("", name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("semaphore"),
	}); err != nil {
		return fmt.Errorf("unable to publish semaphore: %s", err)
	}

	return nil
}

func (c *Client) shouldCreateExclusiveQueue(name string) (bool, int) {
	channel, err := c.connection.Channel()
	if err != nil {
		return false, 0
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}
	}()

	queue, err := channel.QueueInspect(name)
	if err != nil {
		return true, 0
	}

	return false, queue.Messages
}

// Exclusive get an exclusive lock from given queue during duration
func (c *Client) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (err error) {
	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, fmt.Errorf("unable to close channel: %s", err))
		}
	}()

	var message amqp.Delivery
	var acquired bool
	if message, acquired, err = channel.Get(name, false); err != nil {
		return fmt.Errorf("unable to get semaphore: %s", err)
	} else if !acquired {
		return nil
	}

	defer func() {
		if nackErr := message.Nack(false, true); nackErr != nil {
			err = model.WrapError(err, fmt.Errorf("unable to nack message: %s", err))
		}
	}()

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return action(actionCtx)
}

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
	channel, err := c.connection.Channel()
	if err != nil {
		return fmt.Errorf("unable to open channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}
	}()

	var queue amqp.Queue
	queue, err = channel.QueueInspect(name)
	if err != nil {
		if _, err = c.channel.QueueDeclare(name, true, false, false, false, nil); err != nil {
			return fmt.Errorf("unable to declare queue: %s", err)
		}
	} else if queue.Messages > 0 {
		return nil
	}

	if err = c.channel.Publish("", name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("semaphore"),
	}); err != nil {
		return fmt.Errorf("unable to publish semaphore: %s", err)
	}

	return nil
}

// Exclusive get an exclusive lock from given queue during duration
func (c *Client) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) error {
	message, acquired, err := c.channel.Get(name, false)
	if err != nil {
		return fmt.Errorf("unable to get semaphore: %s", err)
	} else if !acquired {
		return nil
	}

	defer c.Reject(message, true)

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return action(actionCtx)
}

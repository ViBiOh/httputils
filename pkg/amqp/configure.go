package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

// Consumer configures client for consumming from given queue, bind to given exchange, and return delayed Exchange name to publish
func (c *Client) Consumer(queueName, routingKey, exchangeName string, exclusive bool, dlExchange string) error {
	channel, err := c.createChannel()
	if err != nil {
		return fmt.Errorf("unable to create channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, fmt.Errorf("unable to close channel: %s", closeErr))
		}
	}()

	var args map[string]interface{}
	if len(dlExchange) != 0 {
		args = map[string]interface{}{
			"x-dead-letter-exchange": dlExchange,
		}
	}

	queue, err := channel.QueueDeclare(queueName, true, false, exclusive, false, args)
	if err != nil {
		return fmt.Errorf("unable to declare queue: %s", err)
	}

	if err := channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil); err != nil {
		return fmt.Errorf("unable to bind queue `%s` to `%s`: %s", queue.Name, exchangeName, err)
	}

	return nil
}

// DelayedExchange configures dead-letter exchange with given ttl
func (c *Client) DelayedExchange(queueName, exchangeName string, retryDelay time.Duration) (string, error) {
	channel, err := c.createChannel()
	if err != nil {
		return "", fmt.Errorf("unable to create channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, fmt.Errorf("unable to close channel: %s", closeErr))
		}
	}()

	delayExchange := fmt.Sprintf("%s-%s", exchangeName, retryDelay)

	if err := declareExchange(channel, delayExchange, "direct", nil); err != nil {
		return "", fmt.Errorf("unable to declare dead-letter exchange: %s", delayExchange)
	}

	delayQueue := fmt.Sprintf("%s-%s", queueName, retryDelay)

	if _, err = channel.QueueDeclare(delayQueue, true, false, false, false, map[string]interface{}{
		"x-dead-letter-exchange": exchangeName,
		"x-message-ttl":          retryDelay.Milliseconds(),
	}); err != nil {
		return "", fmt.Errorf("unable to declare dead-letter queue: %s", delayExchange)
	}

	if err = channel.QueueBind(delayQueue, "", delayExchange, false, nil); err != nil {
		return "", fmt.Errorf("unable to bind dead-letter queue: %s", delayExchange)
	}

	return delayExchange, nil
}

// Publisher configures client for publishing to given exchange
func (c *Client) Publisher(exchangeName, exchangeType string, args amqp.Table) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	channel, err := c.connection.Channel()
	if err != nil {
		return fmt.Errorf("unable to create channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}
	}()

	return declareExchange(channel, exchangeName, exchangeType, args)
}

func declareExchange(channel *amqp.Channel, exchangeName, exchangeType string, args amqp.Table) error {
	if err := channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, args); err != nil {
		return fmt.Errorf("unable to declare exchange `%s`: %s", exchangeName, err)
	}

	return nil
}

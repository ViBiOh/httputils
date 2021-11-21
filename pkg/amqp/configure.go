package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

// Consumer configures client for consumming from given queue, bind to given exchange, and return delayed Exchange name to publish
func (c *Client) Consumer(queueName, routingKey, exchangeName string, exclusive bool, retryDelay time.Duration) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	channel, err := c.connection.Channel()
	if err != nil {
		return "", fmt.Errorf("unable to create channel: %s", err)
	}

	defer func() {
		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}
	}()

	queue, err := channel.QueueDeclare(queueName, true, false, exclusive, false, nil)
	if err != nil {
		return "", fmt.Errorf("unable to declare queue: %s", err)
	}

	if err := channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("unable to bind queue `%s` to `%s`: %s", queue.Name, exchangeName, err)
	}

	var delayExchange string
	if retryDelay != 0 {
		delayExchange = exchangeName + "-delay"

		if err = declareExchange(channel, delayExchange, "direct", map[string]interface{}{
			"x-dead-letter-exchange": exchangeName,
			"x-message-ttl":          retryDelay.Milliseconds(),
		}); err != nil {
			return "", fmt.Errorf("unable to declare delayed exchange: %s", delayExchange)
		}
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

package amqp

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Consumer configures client for consumming from given queue, bind to given exchange, and return delayed Exchange name to publish
func (c *Client) Consumer(queueName, routingKey, exchangeName string, exclusive bool, dlExchange string) (err error) {
	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return err
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	var args map[string]any
	if len(dlExchange) != 0 {
		args = map[string]any{
			"x-dead-letter-exchange":    dlExchange,
			"x-dead-letter-routing-key": routingKey,
		}
	}

	var queue amqp.Queue
	queue, err = channel.QueueDeclare(queueName, true, false, exclusive, false, args)
	if err != nil {
		return fmt.Errorf("unable to declare queue: %s", err)
	}

	if err = channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil); err != nil {
		return fmt.Errorf("unable to bind queue `%s` to `%s`: %s", queue.Name, exchangeName, err)
	}

	return nil
}

// DelayedExchange configures dead-letter exchange with given ttl
func (c *Client) DelayedExchange(queueName, exchangeName, routingKey string, retryDelay time.Duration) (delayExchange string, err error) {
	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return "", err
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	delayExchange = fmt.Sprintf("%s-delay", exchangeName)

	if err = declareExchange(channel, delayExchange, "direct", nil); err != nil {
		return "", fmt.Errorf("unable to declare dead-letter exchange: %s", delayExchange)
	}

	delayQueue := fmt.Sprintf("%s-delay", queueName)

	if _, err = channel.QueueDeclare(delayQueue, true, false, false, false, map[string]any{
		"x-dead-letter-exchange":    exchangeName,
		"x-dead-letter-routing-key": routingKey,
		"x-message-ttl":             retryDelay.Milliseconds(),
	}); err != nil {
		return "", fmt.Errorf("unable to declare dead-letter queue: %s", delayExchange)
	}

	if err = channel.QueueBind(delayQueue, routingKey, delayExchange, false, nil); err != nil {
		return "", fmt.Errorf("unable to bind dead-letter queue: %s", delayExchange)
	}

	return delayExchange, nil
}

// Publisher configures client for publishing to given exchange
func (c *Client) Publisher(exchangeName, exchangeType string, args amqp.Table) (err error) {
	var channel *amqp.Channel
	channel, err = c.createChannel()
	if err != nil {
		return err
	}

	defer func() {
		err = closeChannel(err, channel)
	}()

	return declareExchange(channel, exchangeName, exchangeType, args)
}

func declareExchange(channel *amqp.Channel, exchangeName, exchangeType string, args amqp.Table) error {
	if err := channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, args); err != nil {
		return fmt.Errorf("unable to declare exchange `%s`: %s", exchangeName, err)
	}

	return nil
}

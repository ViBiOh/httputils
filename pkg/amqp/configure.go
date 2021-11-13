package amqp

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Consumer configures client for consumming from given queue, bind to given exchange, and return delayed Exchange name to publish
func (a *Client) Consumer(queueName, routingKey, exchangeName string, retryDelay time.Duration) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	queue, err := a.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return "", fmt.Errorf("unable to declare queue: %s", err)
	}

	if err := a.channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("unable to bind queue `%s` to `%s`: %s", queue.Name, exchangeName, err)
	}

	var delayExchange string
	if retryDelay != 0 {
		delayExchange := exchangeName + "-delay"

		err := a.declareExchange(delayExchange, "direct", map[string]interface{}{
			"x-dead-letter-exchange": exchangeName,
			"x-message-ttl":          retryDelay.Milliseconds(),
		}, false)
		if err != nil {
			return "", fmt.Errorf("unable to declare delayed exchange: %s", delayExchange)
		}
	}

	return delayExchange, nil
}

// Publisher configures client for publishing to given exchange
func (a *Client) Publisher(exchangeName, exchangeType string, args amqp.Table) error {
	return a.declareExchange(exchangeName, exchangeType, args, true)
}

func (a *Client) declareExchange(exchangeName, exchangeType string, args amqp.Table, lock bool) error {
	if lock {
		a.mutex.RLock()
		defer a.mutex.RUnlock()
	}

	if err := a.channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, args); err != nil {
		return fmt.Errorf("unable to declare exchange `%s`: %s", exchangeName, err)
	}

	return nil
}

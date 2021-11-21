package amqphandler

import (
	"errors"
	"fmt"

	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// ErrNoDeathCount occurs when no death count is found in message
var ErrNoDeathCount = errors.New("no death count")

// Retry a message if possible on error
func (a App) Retry(log logger.Provider, message amqp.Delivery) error {
	if a.retry {
		count, err := GetDeathCount(message)
		if err != nil && !errors.Is(err, ErrNoDeathCount) {
			a.amqpClient.Reject(message, false)
			return fmt.Errorf("unable to get death count from message: %s", err)
		}

		if count < a.maxRetry {
			if err := a.amqpClient.Publish(amqpclient.ConvertDeliveryToPublishing(message), a.delayExchange, a.routingKey); err != nil {
				a.amqpClient.Reject(message, true)
				return fmt.Errorf("unable to delay message: %s", err)
			}

			log.Info("message has been delayed in `%s`", a.delayExchange)
		}
	}

	a.amqpClient.Ack(message)
	return nil
}

// GetDeathCount of a message
func GetDeathCount(message amqp.Delivery) (int64, error) {
	table := message.Headers

	rawDeath := table["x-death"]

	death, ok := rawDeath.([]interface{})
	if !ok {
		return 0, fmt.Errorf("`x-death` header in not an array: %w", ErrNoDeathCount)
	}

	if len(death) == 0 {
		return 0, fmt.Errorf("`x-death` is an empty array")
	}

	deathData, ok := death[0].(amqp.Table)
	if !ok {
		return 0, fmt.Errorf("`x-death` datas are not a map")
	}

	count, ok := deathData["count"].(int64)
	if !ok {
		return 0, fmt.Errorf("`count` is not an int")
	}

	return count, nil
}

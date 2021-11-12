package amqp

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

// DeliveryStatus on error
type DeliveryStatus int

const (
	// DeliveryRejected when message is dropped
	DeliveryRejected DeliveryStatus = iota
	// DeliveryDelayed when message is sent in delay queue
	DeliveryDelayed
)

// ErrNoDeathCount occurs when no death count is found in message
var ErrNoDeathCount = errors.New("no death count")

// Retry a message if possible on error
func (c *Client) Retry(message amqp.Delivery, maxRetry int64, delayExchange string) (DeliveryStatus, error) {
	status := DeliveryRejected

	if maxRetry != 0 {
		count, err := GetDeathCount(message)
		if err != nil && !errors.Is(err, ErrNoDeathCount) {
			c.Reject(message, false)
			return DeliveryRejected, fmt.Errorf("unable to get death count from message: %s", err)
		}

		if count < maxRetry && len(delayExchange) > 0 {
			if err := c.Publish(ConvertDeliveryToPublishing(message), delayExchange); err != nil {
				c.Reject(message, true)
				return DeliveryRejected, fmt.Errorf("unable to delay message: %s", err)
			}

			status = DeliveryDelayed
		}
	}

	c.Ack(message)
	return status, nil
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

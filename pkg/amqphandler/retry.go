package amqphandler

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

// ErrNoDeathCount occurs when no death count is found in message.
var ErrNoDeathCount = errors.New("no death count")

// Retry a message if possible on error.
func (a App) Retry(message amqp.Delivery) error {
	count, err := GetDeathCount(message)
	if err != nil && !errors.Is(err, ErrNoDeathCount) {
		return fmt.Errorf("get death count from message: %w", err)
	}

	if count >= a.maxRetry {
		return message.Ack(false)
	}

	return message.Nack(false, false)
}

// GetDeathCount of a message.
func GetDeathCount(message amqp.Delivery) (int64, error) {
	table := message.Headers

	rawDeath := table["x-death"]

	death, ok := rawDeath.([]any)
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

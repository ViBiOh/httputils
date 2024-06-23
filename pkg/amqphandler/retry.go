package amqphandler

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrNoDeathCount = errors.New("no death count")

func (s *Service) Retry(message amqp.Delivery) error {
	count, err := GetDeathCount(message)
	if err != nil && !errors.Is(err, ErrNoDeathCount) {
		return fmt.Errorf("get death count from message: %w", err)
	}

	if count >= s.maxRetry {
		return message.Ack(false)
	}

	return message.Nack(false, false)
}

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

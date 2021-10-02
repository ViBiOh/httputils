package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// Listen listens to configured queue
func (a *Client) Listen(queue string) (string, <-chan amqp.Delivery, error) {
	name, reconnect, err := a.getListener()
	if err != nil {
		return "", nil, fmt.Errorf("unable to get listener name: %s", err)
	}

	messages, err := a.listen(name, queue)
	if err != nil {
		return "", nil, err
	}

	forward := make(chan amqp.Delivery)

	go a.forward(name, queue, reconnect, messages, forward)

	return name, forward, nil
}

// StopListener cancel consumer listening
func (a *Client) StopListener(consumer string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.cancelConsumer(consumer)
	a.removeListener(consumer)
}

func (a *Client) listen(name, queue string) (<-chan amqp.Delivery, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	messages, err := a.channel.Consume(queue, name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to consume queue: %s", err)
	}

	return messages, nil
}

func (a *Client) forward(name, queue string, reconnect <-chan bool, input <-chan amqp.Delivery, output chan<- amqp.Delivery) {
	defer close(output)

forward:
	for delivery := range input {
		a.increase("consumed")
		output <- delivery
	}

	if _, ok := <-reconnect; !ok {
		return
	}

reconnect:
	messages, err := a.listen(name, queue)
	if err != nil {
		logger.Error("unable to reopen listener: %s", err)

		logger.Info("Waiting 30 seconds before attempting to listen again...")
		time.Sleep(time.Second * 30)
		goto reconnect
	}

	input = messages
	goto forward
}

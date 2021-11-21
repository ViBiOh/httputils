package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// Listen listens to configured queue
func (c *Client) Listen(queue string) (string, <-chan amqp.Delivery, error) {
	name, reconnect, err := c.getListener()
	if err != nil {
		return "", nil, fmt.Errorf("unable to get listener name for queue `%s`: %s", queue, err)
	}

	messages, err := c.listen(name, queue)
	if err != nil {
		return "", nil, err
	}

	forward := make(chan amqp.Delivery)

	go c.forward(name, queue, reconnect, messages, forward)

	return name, forward, nil
}

// StopListener cancel consumer listening
func (c *Client) StopListener(consumer string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error
	if err = c.cancelConsumer(consumer); err != nil {
		err = fmt.Errorf("unable to cancel consumer: %s", err)
	}

	c.removeListener(consumer)
	return err
}

func (c *Client) listen(name, queue string) (<-chan amqp.Delivery, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	messages, err := c.channel.Consume(queue, name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to consume queue: %s", err)
	}

	return messages, nil
}

func (c *Client) forward(name, queue string, reconnect <-chan bool, input <-chan amqp.Delivery, output chan<- amqp.Delivery) {
	defer close(output)

forward:
	for delivery := range input {
		c.increase("consumed")
		output <- delivery
	}

	if _, ok := <-reconnect; !ok {
		return
	}

reconnect:
	messages, err := c.listen(name, queue)
	if err != nil {
		logger.Error("unable to reopen listener: %s", err)

		logger.Info("Waiting 30 seconds before attempting to listen again...")
		time.Sleep(time.Second * 30)
		goto reconnect
	}

	input = messages
	goto forward
}

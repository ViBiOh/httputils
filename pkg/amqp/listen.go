package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

// QueueResolver return the name of the queue to listen
type QueueResolver func() (string, error)

// Listen listens to configured queue
func (c *Client) Listen(queueResolver QueueResolver) (string, <-chan amqp.Delivery, error) {
	queueName, err := queueResolver()
	if err != nil {
		return "", nil, fmt.Errorf("unable to get queue name: %s", err)
	}

	listener, err := c.getListener()
	if err != nil {
		return "", nil, fmt.Errorf("unable to get listener name for queue `%s`: %s", queueName, err)
	}

	messages, err := c.listen(listener, queueName)
	if err != nil {
		return "", nil, err
	}

	forward := make(chan amqp.Delivery)
	go c.forward(listener, queueResolver, messages, forward)

	return listener.name, forward, nil
}

// StopListener cancel consumer listening
func (c *Client) StopListener(consumer string) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	listener := c.listeners[consumer]
	if listener == nil {
		return nil
	}

	if cancelErr := listener.cancel(); cancelErr != nil {
		err = fmt.Errorf("unable to cancel listener: %s", err)
	}

	if closeErr := listener.close(); closeErr != nil {
		err = model.WrapError(err, fmt.Errorf("unable to close listener: %s", closeErr))
	}

	c.removeListener(consumer)
	return err
}

func (c *Client) listen(listener *listener, queue string) (<-chan amqp.Delivery, error) {
	listener.RLock()
	defer listener.RUnlock()

	messages, err := listener.channel.Consume(queue, listener.name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to consume queue: %s", err)
	}

	return messages, nil
}

func (c *Client) forward(listener *listener, queueResolver QueueResolver, input <-chan amqp.Delivery, output chan<- amqp.Delivery) {
	defer close(listener.done)
	defer close(output)

forward:
	for delivery := range input {
		c.increase("consumed")
		output <- delivery
	}

	if _, ok := <-listener.reconnect; !ok {
		return
	}

reconnect:
	log := logger.WithField("name", listener.name)

	if queueName, err := queueResolver(); err != nil {
		log.Error("unable to get queue name on reopen: %s", err)
	} else if messages, err := c.listen(listener, queueName); err != nil {
		log.Error("unable to reopen listener: %s", err)
	} else {
		log.Info("Listen restarted.")
		input = messages
		goto forward
	}

	log.Info("Waiting 30 seconds before attempting to listen again...")
	time.Sleep(time.Second * 30)
	goto reconnect
}

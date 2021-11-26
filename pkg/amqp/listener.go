package amqp

import (
	"fmt"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/uuid"
	"github.com/streadway/amqp"
)

type listener struct {
	reconnect chan bool
	done      chan struct{}
	channel   *amqp.Channel
	name      string
	sync.RWMutex
}

func (c *Client) getListener() (*listener, error) {
	listener, err := c.createListener()
	if err != nil {
		return listener, fmt.Errorf("unable to create listener: %s", err)
	}

	c.RLock()
	defer c.RUnlock()

	if err = listener.createChannel(c.connection); err != nil {
		return listener, err
	}

	return listener, nil
}

func (c *Client) createListener() (*listener, error) {
	c.Lock()
	defer c.Unlock()

	var output listener

identity:
	var err error
	output.name, err = uuid.New()
	if err != nil {
		return &output, fmt.Errorf("unable to generate uuid: %s", err)
	}

	if c.listeners[output.name] != nil {
		goto identity
	}

	output.reconnect = make(chan bool, 1)
	output.done = make(chan struct{})

	c.listeners[output.name] = &output
	c.increaseConnection("listener")

	return &output, nil
}

func (l *listener) createChannel(connection Connection) (err error) {
	l.Lock()
	defer l.Unlock()

	if l.channel, err = createChannel(connection); err != nil {
		err = fmt.Errorf("unable to create channel: %s", err)
	}

	return
}

func (l *listener) cancel() error {
	l.RLock()
	defer l.RUnlock()

	close(l.reconnect)
	<-l.reconnect // drain eventually

	return l.channel.Cancel(l.name, false)
}

func (l *listener) close() error {
	l.RLock()
	defer l.RUnlock()

	<-l.done

	return l.channel.Close()
}

func (c *Client) removeListener(name string) {
	listener := c.listeners[name]
	if listener == nil {
		return
	}

	delete(c.listeners, name)
}

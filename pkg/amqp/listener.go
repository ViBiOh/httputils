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
	prefetch  int
	sync.RWMutex
}

func (c *Client) getListener() (*listener, error) {
	listener, err := c.createListener(c.prefetch)
	if err != nil {
		return listener, fmt.Errorf("create listener: %w", err)
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if err = listener.createChannel(c.connection); err != nil {
		return listener, err
	}

	return listener, nil
}

func (c *Client) createListener(prefetch int) (*listener, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var output listener

identity:
	var err error
	output.name, err = uuid.New()
	if err != nil {
		return &output, fmt.Errorf("generate uuid: %w", err)
	}

	if c.listeners[output.name] != nil {
		goto identity
	}

	output.reconnect = make(chan bool, 1)
	output.done = make(chan struct{})
	output.prefetch = prefetch

	c.listeners[output.name] = &output

	if c.listenerMetric != nil {
		c.listenerMetric.Inc()
	}

	return &output, nil
}

func (l *listener) createChannel(connection Connection) (err error) {
	l.Lock()
	defer l.Unlock()

	if l.channel, err = createChannel(connection, l.prefetch); err != nil {
		err = fmt.Errorf("create channel: %w", err)
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
	if listener := c.listeners[name]; listener == nil {
		return
	}

	delete(c.listeners, name)

	if c.listenerMetric != nil {
		c.listenerMetric.Dec()
	}
}

package amqp

import (
	"context"
	"fmt"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/id"
	amqp "github.com/rabbitmq/amqp091-go"
)

type listener struct {
	reconnect chan bool
	done      chan struct{}
	channel   *amqp.Channel
	name      string
	prefetch  int
	sync.RWMutex
}

func (c *Client) getListener(ctx context.Context) (*listener, error) {
	listener := c.createListener(ctx, c.prefetch)

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if err := listener.createChannel(c.connection); err != nil {
		return listener, err
	}

	return listener, nil
}

func (c *Client) createListener(ctx context.Context, prefetch int) *listener {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var output listener

identity:
	output.name = id.New()
	if c.listeners[output.name] != nil {
		goto identity
	}

	output.reconnect = make(chan bool, 1)
	output.done = make(chan struct{})
	output.prefetch = prefetch

	c.listeners[output.name] = &output

	if c.listenerMetric != nil {
		c.listenerMetric.Add(ctx, 1)
	}

	return &output
}

func (l *listener) createChannel(connection Connection) (err error) {
	l.Lock()
	defer l.Unlock()

	if l.channel, err = createChannel(connection, l.prefetch); err != nil {
		err = fmt.Errorf("create channel: %w", err)
	}

	return err
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

func (c *Client) removeListener(ctx context.Context, name string) {
	if listener := c.listeners[name]; listener == nil {
		return
	}

	delete(c.listeners, name)

	if c.listenerMetric != nil {
		c.listenerMetric.Add(ctx, -1)
	}
}

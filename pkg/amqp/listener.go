package amqp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/id"
	amqp "github.com/rabbitmq/amqp091-go"
)

type listener struct {
	cancelErr error
	closeErr  error
	reconnect chan bool
	done      chan struct{}
	stop      chan struct{}
	channel   *amqp.Channel
	name      string
	prefetch  int
	sync.RWMutex
	cancelOnce sync.Once
	closeOnce  sync.Once
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
	output.stop = make(chan struct{})
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
	l.cancelOnce.Do(func() {
		close(l.stop)

		l.RLock()
		defer l.RUnlock()

		close(l.reconnect)
		<-l.reconnect // drop a reconnection already pending

		if l.channel != nil {
			l.cancelErr = l.channel.Cancel(l.name, false)
		}
	})

	return l.cancelErr
}

func (l *listener) close() error {
	l.closeOnce.Do(func() {
		<-l.done

		l.RLock()
		defer l.RUnlock()

		if l.channel != nil {
			l.closeErr = l.channel.Close()
		}
	})

	return l.closeErr
}

func (l *listener) shutdown() (err error) {
	if cancelErr := l.cancel(); cancelErr != nil {
		err = fmt.Errorf("cancel listener: %w", cancelErr)
	}

	if closeErr := l.close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close listener: %w", closeErr))
	}

	return err
}

func (l *listener) signalReconnect() {
	select {
	case l.reconnect <- true:
	default:
	}
}

func (l *listener) closeChannel() error {
	l.Lock()
	defer l.Unlock()

	if l.channel == nil {
		return nil
	}

	err := l.channel.Close()
	l.channel = nil

	return err
}

func (c *Client) discardListener(ctx context.Context, item *listener) {
	c.mutex.Lock()
	c.removeListener(ctx, item.name)
	c.mutex.Unlock()

	if err := item.closeChannel(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "close discarded listener channel", slog.String("name", item.name), slog.Any("error", err))
	}
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

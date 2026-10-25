package amqp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

func (c *Client) Close(ctx context.Context) {
	if c == nil {
		return
	}

	for _, listener := range c.detachListeners(ctx) {
		if err := listener.shutdown(); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "close listener", slog.String("name", listener.name), slog.Any("error", err))
		}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closeChannel(ctx)
	c.closeConnection(ctx)
}

// Listeners are shut down after being detached, without the client lock, because their forwarder may need that lock to finish.
func (c *Client) detachListeners(ctx context.Context) []*listener {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true

	listeners := make([]*listener, 0, len(c.listeners))

	for name, listener := range c.listeners {
		listeners = append(listeners, listener)

		c.removeListener(ctx, name)
	}

	return listeners
}

func (c *Client) reconnect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	newConnection, newChannel, err := connect(c.uri, c.prefetch, c.onDisconnect)
	if err != nil {
		return fmt.Errorf("reconnect to amqp: %w", err)
	}

	c.connection = newConnection
	c.channel = newChannel
	c.vhost = newConnection.Config.Vhost

	slog.InfoContext(ctx, "Connection reopened.")

	go c.reconnectListeners(ctx)

	return nil
}

func (c *Client) reconnectListeners(ctx context.Context) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.connection == nil {
		return
	}

	for _, listener := range c.listeners {
		if err := listener.createChannel(c.connection); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "recreate channel", slog.String("name", listener.name), slog.Any("error", err))
		}

		listener.signalReconnect()
	}
}

func (c *Client) closeChannel(ctx context.Context) {
	if c.channel == nil {
		return
	}

	slog.InfoContext(ctx, "Closing AMQP channel")
	loggedClose(ctx, c.channel)

	c.channel = nil
}

func (c *Client) closeConnection(ctx context.Context) {
	if c.connection == nil {
		return
	}

	if c.connection.IsClosed() {
		c.connection = nil

		return
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Closing AMQP connection", slog.String("vhost", c.Vhost()))
	loggedClose(ctx, c.connection)

	c.connection = nil
}

func loggedClose(ctx context.Context, closer io.Closer) {
	if err := closer.Close(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "close", slog.Any("error", err))
	}
}

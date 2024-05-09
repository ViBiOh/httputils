package amqp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

func (c *Client) Close(ctx context.Context) {
	if c == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error

	if err = c.cancelListeners(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "cancel listeners", slog.Any("error", err))
	}

	if err = c.closeListeners(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "close listeners", slog.Any("error", err))
	}

	c.closeChannel(ctx)
	c.closeConnection(ctx)
}

func (c *Client) reconnect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newConnection, newChannel, err := connect(c.uri, c.prefetch, c.onDisconnect)
	if err != nil {
		return fmt.Errorf("reconnect to amqp: %w", err)
	}

	c.connection = newConnection
	c.channel = newChannel
	c.vhost = newConnection.Config.Vhost

	slog.Info("Connection reopened.")

	go c.reconnectListeners(ctx)

	return nil
}

func (c *Client) cancelListeners() (err error) {
	for _, listener := range c.listeners {
		if cancelErr := listener.cancel(); cancelErr != nil {
			err = errors.Join(err, fmt.Errorf("cancel listener `%s`: %w", listener.name, cancelErr))
		}
	}

	return err
}

func (c *Client) closeListeners() (err error) {
	for _, listener := range c.listeners {
		if cancelErr := listener.close(); cancelErr != nil {
			err = errors.Join(err, fmt.Errorf("close listener `%s`: %w", listener.name, cancelErr))
		}
	}

	return nil
}

func (c *Client) reconnectListeners(ctx context.Context) {
	for _, item := range c.listeners {
		func(listener *listener) {
			c.mutex.Lock()
			defer c.mutex.Unlock()

			if err := listener.createChannel(c.connection); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "recreate channel", slog.String("name", listener.name), slog.Any("error", err))
			}

			listener.reconnect <- true
		}(item)
	}
}

func (c *Client) closeChannel(ctx context.Context) {
	if c.channel == nil {
		return
	}

	slog.Info("Closing AMQP channel")
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

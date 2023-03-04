package amqp

import (
	"errors"
	"fmt"
	"io"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func (c *Client) Close() {
	if c == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error

	if err = c.cancelListeners(); err != nil {
		logger.Error("cancel listeners: %s", err)
	}

	if err = c.closeListeners(); err != nil {
		logger.Error("close listeners: %s", err)
	}

	c.closeChannel()
	c.closeConnection()
}

func (c *Client) reconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newConnection, newChannel, err := connect(c.uri, c.prefetch, c.onDisconnect)
	if err != nil {
		return fmt.Errorf("reconnect to amqp: %w", err)
	}

	c.connection = newConnection
	c.channel = newChannel
	c.vhost = newConnection.Config.Vhost

	logger.Info("Connection reopened.")

	go c.reconnectListeners()

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

func (c *Client) reconnectListeners() {
	for _, item := range c.listeners {
		func(listener *listener) {
			c.mutex.Lock()
			defer c.mutex.Unlock()

			if err := listener.createChannel(c.connection); err != nil {
				logger.WithField("name", listener.name).Error("recreate channel: %s", err)
			}

			listener.reconnect <- true
		}(item)
	}
}

func (c *Client) closeChannel() {
	if c.channel == nil {
		return
	}

	logger.Info("Closing AMQP channel")
	loggedClose(c.channel)

	c.channel = nil
}

func (c *Client) closeConnection() {
	if c.connection == nil {
		return
	}

	if c.connection.IsClosed() {
		c.connection = nil

		return
	}

	logger.WithField("vhost", c.Vhost()).Info("Closing AMQP connection")
	loggedClose(c.connection)

	c.connection = nil
}

func loggedClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		logger.Error("error while closing: %s", err)
	}
}

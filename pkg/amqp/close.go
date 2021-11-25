package amqp

import (
	"fmt"
	"io"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

// Close closes opened ressources
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.cancelListeners(); err != nil {
		logger.Error("unable to cancel listeners: %s", err)
	}
	if err := c.closeListeners(); err != nil {
		logger.Error("unable to close listeners: %s", err)
	}

	c.closeChannel()
	c.closeConnection()
}

func (c *Client) reconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newConnection, newChannel, err := connect(c.uri, c.onDisconnect)
	if err != nil {
		return fmt.Errorf("unable to reconnect to amqp: %s", err)
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
			err = model.WrapError(err, fmt.Errorf("unable to cancel listener `%s`: %s", listener.name, cancelErr))
		}
	}

	return err
}

func (c *Client) closeListeners() (err error) {
	for _, listener := range c.listeners {
		if cancelErr := listener.close(); cancelErr != nil {
			err = model.WrapError(err, fmt.Errorf("unable to close listener `%s`: %s", listener.name, cancelErr))
		}
	}

	return nil
}

func (c *Client) reconnectListeners() {
	for _, item := range c.listeners {
		func(listener *listener) {
			listener.Lock()
			defer listener.Unlock()

			listener.channel = nil

			if channel, err := c.createChannel(); err != nil {
				logger.WithField("name", listener.name).Error("unable to recreate channel: %s", err)
			} else {
				listener.channel = channel
				listener.reconnect <- true
			}
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

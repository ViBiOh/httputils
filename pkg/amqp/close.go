package amqp

import (
	"fmt"
	"io"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Close closes opened ressources
func (c *Client) Close() {
	if err := c.close(false); err != nil {
		logger.Error("unable to close: %s", err)
	}
}

func (c *Client) close(reconnect bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for name := range c.listeners {
		if err := c.cancelConsumer(name); err != nil {
			logger.WithField("name", name).Error("unable to cancel consumer: %s", err)
		}
	}

	c.closeChannel()
	c.closeConnection()

	if !reconnect {
		c.closeListeners()
		return nil
	}

	newConnection, newChannel, err := connect(c.uri, c.onDisconnect)
	if err != nil {
		return fmt.Errorf("unable to reconnect to amqp: %s", err)
	}

	c.connection = newConnection
	c.channel = newChannel
	c.vhost = newConnection.Config.Vhost

	logger.Info("Connection reopened.")

	go c.notifyListeners()

	return nil
}

func (c *Client) cancelConsumer(consumer string) error {
	if err := c.channel.Cancel(consumer, false); err != nil {
		return fmt.Errorf("unable to cancel channel for consumer: %s", err)
	}

	return nil
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

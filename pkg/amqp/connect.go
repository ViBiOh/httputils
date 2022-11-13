package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

func connect(uri string, prefetch int, onDisconnect func()) (*amqp.Connection, *amqp.Channel, error) {
	logger.Info("Dialing AMQP with 10 seconds timeout...")

	connection, err := amqp.DialConfig(uri, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial:      amqp.DefaultDial(10 * time.Second),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("connect to amqp: %w", err)
	}

	channel, err := createChannel(connection, prefetch)
	if err != nil {
		err := fmt.Errorf("create channel: %w", err)

		if closeErr := connection.Close(); closeErr != nil {
			err = model.WrapError(err, fmt.Errorf("close connection: %w", closeErr))
		}

		return nil, nil, err
	}

	go func() {
		log := logger.WithField("addr", connection.LocalAddr().String())
		log.Info("Start listening close connection notifications")

		for range connection.NotifyClose(make(chan *amqp.Error)) {
			log.Warn("Connection closed, trying to reconnect...")
			onDisconnect()
		}

		log.Info("End listening close connection notifications")
	}()

	return connection, channel, nil
}

func createChannel(connection Connection, prefetch int) (channel *amqp.Channel, err error) {
	defer func() {
		if channel == nil || err == nil {
			return
		}

		err = closeChannel(err, channel)
	}()

	channel, err = connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err = channel.Qos(prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("configure QoS on channel: %w", err)
	}

	return channel, nil
}

func (c *Client) onDisconnect() {
	for {
		if c.reconnectMetric != nil {
			c.reconnectMetric.Inc()
		}

		if err := c.reconnect(); err != nil {
			logger.Error("reconnect: %s", err)

			logger.Info("Waiting one minute before attempting to reconnect again...")
			time.Sleep(time.Minute)
		} else {
			return
		}
	}
}

func (c *Client) createChannel() (channel *amqp.Channel, err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	channel, err = createChannel(c.connection, c.prefetch)
	if err != nil {
		err = fmt.Errorf("create channel: %w", err)
	}

	return
}

func closeChannel(err error, channel *amqp.Channel) error {
	if closeErr := channel.Close(); closeErr != nil {
		return model.WrapError(err, fmt.Errorf("close channel: %w", closeErr))
	}

	return err
}

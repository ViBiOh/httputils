package amqp

import (
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/streadway/amqp"
)

func connect(uri string, onDisconnect func()) (*amqp.Connection, *amqp.Channel, error) {
	logger.Info("Dialing AMQP with 10 seconds timeout...")

	connection, err := amqp.DialConfig(uri, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial:      amqp.DefaultDial(10 * time.Second),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to amqp: %s", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		err := fmt.Errorf("unable to open communication channel: %s", err)

		if closeErr := connection.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}

		return nil, nil, err
	}

	if err := channel.Qos(1, 0, false); err != nil {
		err := fmt.Errorf("unable to configure QoS on channel: %s", err)

		if closeErr := channel.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}

		if closeErr := connection.Close(); closeErr != nil {
			err = model.WrapError(err, closeErr)
		}

		return nil, nil, err
	}

	go func() {
		log := logger.WithField("addr", connection.LocalAddr().String())
		log.Info("Start listening close connection notifications")
		defer log.Info("End listening close connection notifications")

		for range connection.NotifyClose(make(chan *amqp.Error)) {
			log.Warn("Connection closed, trying to reconnect.")
			onDisconnect()
		}
	}()

	return connection, channel, nil
}

func (a *Client) onDisconnect() {
	for {
		a.increaseConnection("reconnect")

		if err := a.close(true); err != nil {
			logger.Error("unable to reconnect: %s", err)

			logger.Info("Waiting one minute before attempting to reconnect again...")
			time.Sleep(time.Minute)
		} else {
			return
		}
	}
}

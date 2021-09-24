package amqp

import (
	"fmt"
	"io"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Close closes opened ressources
func (a *Client) Close() {
	if err := a.close(false); err != nil {
		logger.Error("unable to close: %s", err)
	}
}

func (a *Client) close(reconnect bool) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for name := range a.listeners {
		a.cancelConsumer(name)
	}

	a.closeChannel()
	a.closeConnection()

	if !reconnect {
		a.closeListeners()
		return nil
	}

	newConnection, newChannel, err := connect(a.uri, a.onDisconnect)
	if err != nil {
		return fmt.Errorf("unable to reconnect to amqp: %s", err)
	}

	a.connection = newConnection
	a.channel = newChannel
	a.vhost = newConnection.Config.Vhost

	logger.Info("Connection reopened.")

	go a.notifyListeners()

	return nil
}

func (a *Client) cancelConsumer(consumer string) {
	log := logger.WithField("consumer", consumer)

	log.Info("Canceling AMQP channel")
	if err := a.channel.Cancel(consumer, false); err != nil {
		log.Error("unable to cancel consumer: %s", err)
	}
}

func (a *Client) closeChannel() {
	if a.channel == nil {
		return
	}

	logger.Info("Closing AMQP channel")
	loggedClose(a.channel)

	a.channel = nil
}

func (a *Client) closeConnection() {
	if a.connection == nil {
		return
	}

	if a.connection.IsClosed() {
		a.connection = nil
		return
	}

	logger.WithField("vhost", a.Vhost()).Info("Closing AMQP connection")
	loggedClose(a.connection)

	a.connection = nil
}

func loggedClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		logger.Error("error while closing: %s", err)
	}
}

package amqp

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// Connection for AMQP
//go:generate mockgen -destination ../mocks/amqp.go -mock_names Connection=AMQPConnection -package mocks github.com/ViBiOh/httputils/v4/pkg/amqp Connection
type Connection interface {
	io.Closer
	IsClosed() bool
}

// Client wraps all object required for AMQP usage
type Client struct {
	channel    *amqp.Channel
	connection Connection
	listeners  map[string]chan bool
	vhost      string
	uri        string
	mutex      sync.RWMutex
}

// New inits AMQP connection, channel and queue
func New(uri string) (*Client, error) {
	if len(uri) == 0 {
		return nil, errors.New("URI is required")
	}

	client := &Client{
		uri: uri,
	}

	connection, channel, err := connect(uri, client.onDisconnect)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to amqp: %s", err)
	}

	client.connection = connection
	client.channel = channel
	client.vhost = connection.Config.Vhost

	logger.WithField("vhost", client.vhost).Info("Connected to AMQP!")

	return client, nil
}

// Publish sends payload to the underlying exchange
func (a *Client) Publish(payload amqp.Publishing, exchange string) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.channel.Publish(exchange, "", false, false, payload)
}

// Ack ack a message with error handling
func (a *Client) Ack(message amqp.Delivery) {
	a.loggerMessageDeliveryAckReject(message, true, false)
}

// Reject reject a message with error handling
func (a *Client) Reject(message amqp.Delivery, requeue bool) {
	a.loggerMessageDeliveryAckReject(message, false, requeue)
}

func (a *Client) loggerMessageDeliveryAckReject(message amqp.Delivery, ack bool, value bool) {
	for {
		var err error

		if ack {
			err = message.Ack(value)
		} else {
			err = message.Reject(value)
		}

		if err == nil {
			return
		}

		if err != amqp.ErrClosed {
			logger.Error("unable to ack/reject message: %s", err)
			return
		}

		logger.Error("unable to ack/reject message due to a closed connection")

		logger.Info("Waiting 30 seconds before attempting to ack/reject message again...")
		time.Sleep(time.Second * 30)

		func() {
			a.mutex.RLock()
			defer a.mutex.RUnlock()

			message.Acknowledger = a.channel
		}()
	}
}

// ConvertDeliveryToPublishing convert a delivery to a publishing, for requeuing
func ConvertDeliveryToPublishing(message amqp.Delivery) amqp.Publishing {
	return amqp.Publishing{
		Headers:         message.Headers,
		ContentType:     message.ContentType,
		ContentEncoding: message.ContentEncoding,
		DeliveryMode:    message.DeliveryMode,
		Priority:        message.Priority,
		CorrelationId:   message.CorrelationId,
		ReplyTo:         message.ReplyTo,
		Expiration:      message.Expiration,
		MessageId:       message.MessageId,
		Timestamp:       message.Timestamp,
		Type:            message.Type,
		UserId:          message.UserId,
		AppId:           message.AppId,
		Body:            message.Body,
	}
}

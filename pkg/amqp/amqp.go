package amqp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// Connection for AMQP
type Connection interface {
	io.Closer
	IsClosed() bool
}

// Client wraps all object required for AMQP usage
type Client struct {
	channel            *amqp.Channel
	connection         Connection
	vhost              string
	clientName         string
	uri                string
	reconnectListeners []chan bool
	mutex              sync.RWMutex
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
			err = fmt.Errorf("%s: %w", err, closeErr)
		}

		return nil, nil, err
	}

	if err := channel.Qos(1, 0, false); err != nil {
		err := fmt.Errorf("unable to configure QoS on channel: %s", err)

		if closeErr := channel.Close(); closeErr != nil {
			err = fmt.Errorf("%s: %w", err, closeErr)
		}

		if closeErr := connection.Close(); closeErr != nil {
			err = fmt.Errorf("%s: %w", err, closeErr)
		}

		return nil, nil, err
	}

	go func() {
		localAddr := connection.LocalAddr()
		logger.Warn("Listening close notifications %s", localAddr)
		defer logger.Warn("Close notifications are over for %s", localAddr)

		for range connection.NotifyClose(make(chan *amqp.Error)) {
			logger.Warn("Connection closed, trying to reconnect.")
			onDisconnect()
		}

	}()

	return connection, channel, nil
}

// ListenReconnect creates a chan notifier with a boolean when doing reconnection
func (a *Client) ListenReconnect() <-chan bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	listener := make(chan bool)
	a.reconnectListeners = append(a.reconnectListeners, listener)

	return listener
}

func (a *Client) closeListeners() {
	for _, listener := range a.reconnectListeners {
		close(listener)
	}
}

func (a *Client) notifyListeners() {
	for _, listener := range a.reconnectListeners {
		listener <- true
	}
}

// Consumer configures client for consumming from given queue, bind to given exchange, and return delayed Exchange name to publish
func (a *Client) Consumer(queueName, topic, exchangeName string, retryDelay time.Duration) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	queue, err := a.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return "", fmt.Errorf("unable to declare queue: %s", err)
	}

	if err := a.channel.QueueBind(queue.Name, topic, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("unable to bind queue `%s` to `%s`: %s", queue.Name, exchangeName, err)
	}

	var delayExchange string
	if retryDelay != 0 {
		delayExchange := exchangeName + "-delay"

		err := a.declareExchange(delayExchange, "direct", map[string]interface{}{
			"x-dead-letter-exchange": exchangeName,
			"x-message-ttl":          retryDelay.Milliseconds(),
		}, false)
		if err != nil {
			return "", fmt.Errorf("unable to declare delayed exchange: %s", delayExchange)
		}
	}

	a.ensureClientName()

	return delayExchange, nil
}

// Publisher configures client for publishing to given exchange
func (a *Client) Publisher(exchangeName, exchangeType string, args amqp.Table) error {
	return a.declareExchange(exchangeName, exchangeType, args, true)
}

// Publisher configures client for publishing to given exchange
func (a *Client) declareExchange(exchangeName, exchangeType string, args amqp.Table, lock bool) error {
	if lock {
		a.mutex.RLock()
		defer a.mutex.RUnlock()
	}

	if err := a.channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, args); err != nil {
		return fmt.Errorf("unable to declare exchange `%s`: %s", exchangeName, err)
	}

	return nil
}

func (a *Client) ensureClientName() {
	if len(a.clientName) != 0 {
		return
	}

	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		logger.Fatal(err)
		a.clientName = "mailer"
	}

	a.clientName = fmt.Sprintf("%x", raw)
}

// Enabled checks if connection is setup
func (a *Client) Enabled() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.connection != nil
}

// Ping checks if connection is live
func (a *Client) Ping() error {
	if !a.Enabled() {
		return nil
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.connection.IsClosed() {
		return errors.New("amqp client closed")
	}

	return nil
}

// ClientName returns client name
func (a *Client) ClientName() string {
	return a.clientName
}

// Vhost returns connection Vhost
func (a *Client) Vhost() string {
	return a.vhost
}

// Publish sends payload to the underlying exchange
func (a *Client) Publish(payload amqp.Publishing, exchange string) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.channel.Publish(exchange, "", false, false, payload)
}

// Listen listens to configured queue
func (a *Client) Listen(queue string) (<-chan amqp.Delivery, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.ensureClientName()

	messages, err := a.channel.Consume(queue, a.clientName, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to consume queue: %s", err)
	}

	return messages, nil
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

// Close closes opened ressources
func (a *Client) Close() {
	if err := a.close(false); err != nil {
		logger.Error("unable to close: %s", err)
	}
}

func (a *Client) close(reconnect bool) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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

func (a *Client) onDisconnect() {
	for {
		if err := a.close(true); err != nil {
			logger.Error("unable to reconnect: %s", err)

			logger.Info("Waiting one minute before attempting to reconnect again...")
			time.Sleep(time.Minute)
		} else {
			return
		}
	}
}

// StopChannel cancel existing channel
func (a *Client) StopChannel() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.channel == nil {
		return
	}

	a.closeChannel()
}

func (a *Client) closeChannel() {
	if a.channel == nil {
		return
	}

	if len(a.clientName) != 0 {
		log := logger.WithField("name", a.clientName)

		log.Info("Canceling AMQP channel")
		if err := a.channel.Cancel(a.clientName, false); err != nil {
			log.Error("unable to cancel consumer: %s", err)
		}
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

package amqphandler

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

// App of package
type App struct {
	amqpClient    *amqpclient.Client
	done          chan struct{}
	handler       func(amqp.Delivery) error
	exchange      string
	delayExchange string
	queue         string
	routingKey    string
	maxRetry      int64
	retryInterval time.Duration
	exclusive     bool
}

// Config of package
type Config struct {
	exchange      *string
	queue         *string
	routingKey    *string
	retryInterval *string
	maxRetry      *uint
	exclusive     *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		exchange:      flags.New(prefix, "amqp", "Exchange").Default("", overrides).Label("Exchange name").ToString(fs),
		queue:         flags.New(prefix, "amqp", "Queue").Default("", overrides).Label("Queue name").ToString(fs),
		exclusive:     flags.New(prefix, "amqp", "Exclusive").Default(false, overrides).Label("Queue exclusive mode (for fanout exchange)").ToBool(fs),
		routingKey:    flags.New(prefix, "amqp", "RoutingKey").Default("", overrides).Label("RoutingKey name").ToString(fs),
		retryInterval: flags.New(prefix, "amqp", "RetryInterval").Default("1h", overrides).Label("Interval duration when send fails").ToString(fs),
		maxRetry:      flags.New(prefix, "amqp", "MaxRetry").Default(3, overrides).Label("Max send retries").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, amqpClient *amqpclient.Client, handler func(amqp.Delivery) error) (App, error) {
	return NewFromString(amqpClient, handler, strings.TrimSpace(*config.exchange), strings.TrimSpace(*config.queue), strings.TrimSpace(*config.routingKey), strings.TrimSpace(*config.retryInterval), *config.exclusive, *config.maxRetry)
}

// NewFromString creates new App from string configuration
func NewFromString(amqpClient *amqpclient.Client, handler func(amqp.Delivery) error, exchange, queue, routingKey, retryInterval string, exclusive bool, maxRetry uint) (App, error) {
	app := App{
		amqpClient: amqpClient,
		exchange:   exchange,
		queue:      queue,
		exclusive:  exclusive,
		routingKey: routingKey,
		done:       make(chan struct{}),
		handler:    handler,
		maxRetry:   int64(maxRetry),
	}

	if app.amqpClient == nil {
		return app, nil
	}

	var err error
	app.retryInterval, err = time.ParseDuration(retryInterval)
	if err != nil {
		return app, fmt.Errorf("unable to parse retry duration: %s", err)
	}

	if app.retryInterval > 0 && app.maxRetry > 0 {
		if len(exchange) == 0 {
			return app, errors.New("no exchange name for delaying retries")
		}

		if app.delayExchange, err = app.amqpClient.DelayedExchange(queue, exchange, routingKey, app.retryInterval); err != nil {
			return app, fmt.Errorf("unable to configure dead-letter exchange: %s", err)
		}
	}

	return app, nil
}

// Done returns the chan used for synchronization
func (a App) Done() <-chan struct{} {
	return a.done
}

// Start amqp handler
func (a App) Start(done <-chan struct{}) {
	defer close(a.done)

	if a.amqpClient == nil {
		return
	}

	init := true
	log := logger.WithField("exchange", a.exchange).WithField("queue", a.queue).WithField("routingKey", a.routingKey).WithField("vhost", a.amqpClient.Vhost())

	consumerName, messages, err := a.amqpClient.Listen(func() (string, error) {
		queueName, err := a.configure(init)
		init = false
		return queueName, err
	}, a.exchange, a.routingKey)
	if err != nil {
		log.Error("unable to listen: %s", err)
		return
	}

	log = log.WithField("name", consumerName)

	go func() {
		<-done
		if err := a.amqpClient.StopListener(consumerName); err != nil {
			log.Error("error while stopping listener: %s", err)
		}
	}()

	log.Info("Start listening messages")
	defer log.Info("End listening messages")

	for message := range messages {
		a.handleMessage(log, message)
	}
}

func (a App) handleMessage(log logger.Provider, message amqp.Delivery) {
	err := a.handler(message)

	if err == nil {
		if err = message.Ack(false); err != nil {
			log.Error("unable to ack message: %s", err)
		}
		return
	}

	log.Error("unable to handle message `%s`: %s", message.Body, err)

	if a.retryInterval > 0 && a.maxRetry > 0 {
		if err = a.Retry(log, message); err == nil {
			return
		}

		log.Error("unable to retry message: %s", err)
	}

	if err = message.Ack(false); err != nil {
		log.Error("unable to ack message to trash it: %s", err)
	}
}

func (a App) configure(init bool) (string, error) {
	if !a.exclusive && !init {
		return a.queue, nil
	}

	queue := a.queue
	if a.exclusive {
		queue = fmt.Sprintf("%s-%s", a.queue, generateIdentityName())
	}

	if err := a.amqpClient.Consumer(queue, a.routingKey, a.exchange, a.exclusive, a.delayExchange); err != nil {
		return "", fmt.Errorf("unable to configure amqp consumer for routingKey `%s` and exchange `%s`: %s", a.routingKey, a.exchange, err)
	}

	return queue, nil
}

func generateIdentityName() string {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		logger.Error("unable to generate identity name: %s", err)
		return "error"
	}

	return fmt.Sprintf("%x", raw)
}

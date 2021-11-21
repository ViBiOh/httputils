package amqphandler

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/sha"
	"github.com/streadway/amqp"
)

// App of package
type App struct {
	amqpClient    *amqpclient.Client
	done          chan struct{}
	handler       func(amqp.Delivery) error
	queue         string
	delayExchange string
	routingKey    string
	maxRetry      int64
	retry         bool
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
		queue:      queue,
		routingKey: routingKey,
		done:       make(chan struct{}),
		handler:    handler,
		maxRetry:   int64(maxRetry),
	}

	if app.amqpClient == nil {
		return app, nil
	}

	retryIntervalDuration, err := time.ParseDuration(retryInterval)
	if err != nil {
		return app, fmt.Errorf("unable to parse retry duration: %s", err)
	}
	app.retry = retryIntervalDuration > 0 && app.maxRetry > 0

	if app.retry && len(exchange) == 0 {
		return app, errors.New("no exchange name for delaying retries")
	}

	if app.delayExchange, err = app.amqpClient.Consumer(app.queue, routingKey, exchange, exclusive, retryIntervalDuration); err != nil {
		return app, fmt.Errorf("unable to configure amqp consumer: %s", err)
	}

	return app, nil
}

// Enabled checks if requirements are met
func (a App) Enabled() bool {
	return a.amqpClient != nil
}

// Done returns the chan used for synchronization
func (a App) Done() <-chan struct{} {
	return a.done
}

// Start amqp handler
func (a App) Start(done <-chan struct{}) {
	defer close(a.done)

	if !a.Enabled() {
		return
	}

	consumerName, messages, err := a.amqpClient.Listen(a.queue)
	if err != nil {
		logger.Error("unable to listen `%s`: %s", a.queue, err)
		return
	}

	go func() {
		<-done
		if err := a.amqpClient.StopListener(consumerName); err != nil {
			logger.WithField("name", consumerName).WithField("queue", a.queue).Error("error while stopping listener: %s", err)
		}
	}()

	log := logger.WithField("queue", a.queue).WithField("name", consumerName).WithField("vhost", a.amqpClient.Vhost())
	log.Info("Start listening messages")
	defer log.Info("End listening messages")

	for message := range messages {
		err := a.handler(message)

		if err == nil {
			a.amqpClient.Ack(message)
			continue
		}

		messageLog := log.WithField("exchange", message.Exchange).WithField("routingKey", message.RoutingKey).WithField("sha", sha.New(message.Body))
		messageLog.Error("unable to handle message: %s", err)

		if err = a.Retry(messageLog, message); err != nil {
			messageLog.Info("unable to retry message: %s", err)
		}
	}
}

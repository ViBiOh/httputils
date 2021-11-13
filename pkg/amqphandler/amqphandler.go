package amqphandler

import (
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
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		exchange:      flags.New(prefix, "amqp", "Exchange").Default("", overrides).Label("Exchange name").ToString(fs),
		queue:         flags.New(prefix, "amqp", "Queue").Default("", overrides).Label("Queue name").ToString(fs),
		routingKey:    flags.New(prefix, "amqp", "RoutingKey").Default("", overrides).Label("RoutingKey name").ToString(fs),
		retryInterval: flags.New(prefix, "amqp", "RetryInterval").Default("1h", nil).Label("Interval duration when send fails").ToString(fs),
		maxRetry:      flags.New(prefix, "amqp", "MaxRetry").Default(3, nil).Label("Max send retries").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, amqpClient *amqpclient.Client, handler func(amqp.Delivery) error) (App, error) {
	return NewFromString(amqpClient, handler, strings.TrimSpace(*config.exchange), strings.TrimSpace(*config.queue), strings.TrimSpace(*config.routingKey), strings.TrimSpace(*config.retryInterval), *config.maxRetry)
}

// NewFromString creates new App from string configuration
func NewFromString(amqpClient *amqpclient.Client, handler func(amqp.Delivery) error, exchange, queue, routingKey, retryInterval string, maxRetry uint) (App, error) {
	app := App{
		amqpClient: amqpClient,
		queue:      queue,
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
	app.retry = retryIntervalDuration != 0 && app.maxRetry > 0

	app.delayExchange, err = app.amqpClient.Consumer(app.queue, routingKey, exchange, retryIntervalDuration)
	if err != nil {
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
		a.amqpClient.StopListener(consumerName)
	}()

	log := logger.WithField("queue", a.queue).WithField("vhost", a.amqpClient.Vhost())
	log.Info("Start listening messages")
	defer log.Info("End listening messages")

	for message := range messages {
		err := a.handler(message)

		if err == nil {
			a.amqpClient.Ack(message)
			continue
		}

		messageSha := sha.New(message.Body)
		log.Error("unable to handle message with sha `%s`: %s", messageSha, err)

		if err = a.Retry(message); err != nil {
			log.Info("unable to retry message with sha `%s`: %s", messageSha, err)
		}
	}
}

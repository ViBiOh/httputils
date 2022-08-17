package amqphandler

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/streadway/amqp"
)

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

type Config struct {
	exchange      *string
	queue         *string
	routingKey    *string
	retryInterval *time.Duration
	maxRetry      *uint
	exclusive     *bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		exchange:      flags.String(fs, prefix, "amqp", "Exchange", "Exchange name", "", overrides),
		queue:         flags.String(fs, prefix, "amqp", "Queue", "Queue name", "", overrides),
		exclusive:     flags.Bool(fs, prefix, "amqp", "Exclusive", "Queue exclusive mode (for fanout exchange)", false, overrides),
		routingKey:    flags.String(fs, prefix, "amqp", "RoutingKey", "RoutingKey name", "", overrides),
		retryInterval: flags.Duration(fs, prefix, "amqp", "RetryInterval", "Interval duration when send fails", time.Hour, overrides),
		maxRetry:      flags.Uint(fs, prefix, "amqp", "MaxRetry", "Max send retries", 3, overrides),
	}
}

func New(config Config, amqpClient *amqpclient.Client, handler func(amqp.Delivery) error) (App, error) {
	return NewFromString(amqpClient, handler, strings.TrimSpace(*config.exchange), strings.TrimSpace(*config.queue), strings.TrimSpace(*config.routingKey), *config.retryInterval, *config.exclusive, *config.maxRetry)
}

// NewFromString creates new App from string configuration.
func NewFromString(amqpClient *amqpclient.Client, handler func(amqp.Delivery) error, exchange, queue, routingKey string, retryInterval time.Duration, exclusive bool, maxRetry uint) (App, error) {
	app := App{
		amqpClient:    amqpClient,
		exchange:      exchange,
		queue:         queue,
		exclusive:     exclusive,
		routingKey:    routingKey,
		retryInterval: retryInterval,
		done:          make(chan struct{}),
		handler:       handler,
		maxRetry:      int64(maxRetry),
	}

	if app.amqpClient == nil {
		return app, nil
	}

	if app.retryInterval > 0 && app.maxRetry > 0 {
		if len(exchange) == 0 {
			return app, errors.New("no exchange name for delaying retries")
		}

		var err error
		if app.delayExchange, err = app.amqpClient.DelayedExchange(queue, exchange, routingKey, app.retryInterval); err != nil {
			return app, fmt.Errorf("configure dead-letter exchange: %w", err)
		}
	}

	return app, nil
}

// Done returns the chan used for synchronization.
func (a App) Done() <-chan struct{} {
	return a.done
}

// Start amqp handler.
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
		log.Error("listen: %s", err)

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
			log.Error("ack message: %s", err)
		}

		return
	}

	log.Error("handle message `%s`: %s", message.Body, err)

	if a.retryInterval > 0 && a.maxRetry > 0 {
		if err = a.Retry(message); err == nil {
			return
		}

		log.Error("retry message: %s", err)
	}

	if err = message.Ack(false); err != nil {
		log.Error("ack message to trash it: %s", err)
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
		return "", fmt.Errorf("configure amqp consumer for routingKey `%s` and exchange `%s`: %w", a.routingKey, a.exchange, err)
	}

	return queue, nil
}

func generateIdentityName() string {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		logger.Error("generate identity name: %s", err)

		return "error"
	}

	return fmt.Sprintf("%x", raw)
}

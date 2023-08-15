package amqphandler

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

type Handler func(context.Context, amqp.Delivery) error

type App struct {
	amqpClient    *amqpclient.Client
	tracer        trace.Tracer
	done          chan struct{}
	handler       Handler
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
		exchange:      flags.New("Exchange", "Exchange name").Prefix(prefix).DocPrefix("amqp").String(fs, "", overrides),
		queue:         flags.New("Queue", "Queue name").Prefix(prefix).DocPrefix("amqp").String(fs, "", overrides),
		exclusive:     flags.New("Exclusive", "Queue exclusive mode (for fanout exchange)").Prefix(prefix).DocPrefix("amqp").Bool(fs, false, overrides),
		routingKey:    flags.New("RoutingKey", "RoutingKey name").Prefix(prefix).DocPrefix("amqp").String(fs, "", overrides),
		retryInterval: flags.New("RetryInterval", "Interval duration when send fails").Prefix(prefix).DocPrefix("amqp").Duration(fs, time.Hour, overrides),
		maxRetry:      flags.New("MaxRetry", "Max send retries").Prefix(prefix).DocPrefix("amqp").Uint(fs, 3, overrides),
	}
}

func New(config Config, amqpClient *amqpclient.Client, tracer trace.Tracer, handler Handler) (*App, error) {
	app := &App{
		amqpClient:    amqpClient,
		tracer:        tracer,
		exchange:      strings.TrimSpace(*config.exchange),
		queue:         strings.TrimSpace(*config.queue),
		exclusive:     *config.exclusive,
		routingKey:    strings.TrimSpace(*config.routingKey),
		retryInterval: *config.retryInterval,
		done:          make(chan struct{}),
		handler:       handler,
		maxRetry:      int64(*config.maxRetry),
	}

	if app.amqpClient == nil {
		return app, nil
	}

	if app.retryInterval > 0 && app.maxRetry > 0 {
		if len(app.exchange) == 0 {
			return app, errors.New("no exchange name for delaying retries")
		}

		var err error
		if app.delayExchange, err = app.amqpClient.DelayedExchange(app.queue, app.exchange, app.routingKey, app.retryInterval); err != nil {
			return app, fmt.Errorf("configure dead-letter exchange: %w", err)
		}
	}

	return app, nil
}

func (a *App) Done() <-chan struct{} {
	return a.done
}

func (a *App) Start(ctx context.Context) {
	defer close(a.done)

	if a.amqpClient == nil {
		return
	}

	init := true
	log := slog.With("exchange", a.exchange).With("queue", a.queue).With("routingKey", a.routingKey).With("vhost", a.amqpClient.Vhost())

	consumerName, messages, err := a.amqpClient.Listen(func() (string, error) {
		queueName, err := a.configure(init)
		init = false

		return queueName, err
	}, a.exchange, a.routingKey)
	if err != nil {
		log.Error("listen", "err", err)

		return
	}

	log = log.With("name", consumerName)

	go func() {
		<-ctx.Done()
		if err := a.amqpClient.StopListener(consumerName); err != nil {
			log.Error("error while stopping listener", "err", err)
		}
	}()

	log.Info("Start listening messages")
	defer log.Info("End listening messages")

	for message := range messages {
		a.handleMessage(ctx, log, message)
	}
}

func (a *App) handleMessage(ctx context.Context, log *slog.Logger, message amqp.Delivery) {
	var err error

	ctx, end := telemetry.StartSpan(ctx, a.tracer, "handle", trace.WithSpanKind(trace.SpanKindConsumer))
	defer end(&err)

	defer recoverer.Error(&err)

	err = a.handler(ctx, message)

	if err == nil {
		if err = message.Ack(false); err != nil {
			log.Error("ack message", "err", err)
		}

		return
	}

	log.Error("handle message", "err", err, "body", string(message.Body))

	if a.retryInterval > 0 && a.maxRetry > 0 {
		if err = a.Retry(message); err == nil {
			return
		}

		log.Error("retry message", "err", err)
	}

	if err = message.Ack(false); err != nil {
		log.Error("ack message to trash it", "err", err)
	}
}

func (a *App) configure(init bool) (string, error) {
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
		slog.Error("generate identity name", "err", err)

		return "error"
	}

	return fmt.Sprintf("%x", raw)
}

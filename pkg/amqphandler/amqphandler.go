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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Handler func(context.Context, amqp.Delivery) error

type Service struct {
	amqpClient    *amqpclient.Client
	tracer        trace.Tracer
	done          chan struct{}
	handler       Handler
	counter       metric.Int64Counter
	exchange      string
	delayExchange string
	queue         string
	routingKey    string
	maxRetry      int64
	retryInterval time.Duration
	exclusive     bool
}

type Config struct {
	Exchange      string
	Queue         string
	RoutingKey    string
	RetryInterval time.Duration
	MaxRetry      uint
	Exclusive     bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("Exchange", "Exchange name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Exchange, "", overrides)
	flags.New("Queue", "Queue name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Queue, "", overrides)
	flags.New("Exclusive", "Queue exclusive mode (for fanout exchange)").Prefix(prefix).DocPrefix("amqp").BoolVar(fs, &config.Exclusive, false, overrides)
	flags.New("RoutingKey", "RoutingKey name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.RoutingKey, "", overrides)
	flags.New("RetryInterval", "Interval duration when send fails").Prefix(prefix).DocPrefix("amqp").DurationVar(fs, &config.RetryInterval, time.Hour, overrides)
	flags.New("MaxRetry", "Max send retries").Prefix(prefix).DocPrefix("amqp").UintVar(fs, &config.MaxRetry, 3, overrides)

	return config
}

func New(config Config, amqpClient *amqpclient.Client, metricProvider metric.MeterProvider, tracerProvider trace.TracerProvider, handler Handler) (*Service, error) {
	service := &Service{
		amqpClient:    amqpClient,
		exchange:      strings.TrimSpace(config.Exchange),
		queue:         strings.TrimSpace(config.Queue),
		exclusive:     config.Exclusive,
		routingKey:    strings.TrimSpace(config.RoutingKey),
		retryInterval: config.RetryInterval,
		done:          make(chan struct{}),
		handler:       handler,
		maxRetry:      int64(config.MaxRetry),
	}

	if service.amqpClient == nil {
		return service, nil
	}

	if service.retryInterval > 0 && service.maxRetry > 0 {
		if len(service.exchange) == 0 {
			return service, errors.New("no exchange name for delaying retries")
		}

		var err error
		if service.delayExchange, err = service.amqpClient.DelayedExchange(service.queue, service.exchange, service.routingKey, service.retryInterval); err != nil {
			return service, fmt.Errorf("configure dead-letter exchange: %w", err)
		}
	}

	if metricProvider != nil {
		meter := metricProvider.Meter("github.com/ViBiOh/httputils/v4/pkg/amqphandler")

		var err error

		service.counter, err = meter.Int64Counter("amqp.message")
		if err != nil {
			return service, fmt.Errorf("create counter: %w", err)
		}
	}

	if tracerProvider != nil {
		service.tracer = tracerProvider.Tracer("amqp_handler")
	}

	return service, nil
}

func (a *Service) Done() <-chan struct{} {
	return a.done
}

func (a *Service) Start(ctx context.Context) {
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

func (a *Service) handleMessage(ctx context.Context, log *slog.Logger, message amqp.Delivery) {
	var err error

	ctx, end := telemetry.StartSpan(ctx, a.tracer, "handle", trace.WithSpanKind(trace.SpanKindConsumer))
	defer end(&err)

	defer recoverer.Error(&err)

	err = a.handler(ctx, message)

	if err == nil {
		a.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "ack"),
			attribute.String("exchange", a.exchange),
			attribute.String("routingKey", a.routingKey),
		))
		if err = message.Ack(false); err != nil {
			log.Error("ack message", "err", err)
		}

		return
	}

	log.Error("handle message", "err", err, "body", string(message.Body))

	if a.retryInterval > 0 && a.maxRetry > 0 {
		a.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "retry"),
			attribute.String("exchange", a.exchange),
			attribute.String("routingKey", a.routingKey),
		))

		if err = a.Retry(message); err == nil {
			return
		}

		log.Error("retry message", "err", err)
	}

	if err = message.Ack(false); err != nil {
		a.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "drop"),
			attribute.String("exchange", a.exchange),
			attribute.String("routingKey", a.routingKey),
		))

		log.Error("ack message to trash it", "err", err)
	}
}

func (a *Service) configure(init bool) (string, error) {
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

package amqphandler

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/ViBiOh/flags"
	amqpclient "github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type Handler func(context.Context, amqp.Delivery) error

type Service struct {
	tracer          trace.Tracer
	counter         metric.Int64Counter
	amqpClient      *amqpclient.Client
	done            chan struct{}
	handler         Handler
	delayExchange   string
	exchange        string
	queue           string
	routingKey      string
	attributes      []attribute.KeyValue
	maxRetry        int64
	retryInterval   time.Duration
	inactiveTimeout time.Duration
	exclusive       bool
}

type Config struct {
	Exchange        string
	Queue           string
	RoutingKey      string
	RetryInterval   time.Duration
	InactiveTimeout time.Duration
	MaxRetry        uint
	Exclusive       bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Exchange", "Exchange name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Exchange, "", overrides)
	flags.New("Queue", "Queue name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Queue, "", overrides)
	flags.New("Exclusive", "Queue exclusive mode (for fanout exchange)").Prefix(prefix).DocPrefix("amqp").BoolVar(fs, &config.Exclusive, false, overrides)
	flags.New("RoutingKey", "RoutingKey name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.RoutingKey, "", overrides)
	flags.New("RetryInterval", "Interval duration when send fails").Prefix(prefix).DocPrefix("amqp").DurationVar(fs, &config.RetryInterval, time.Hour, overrides)
	flags.New("MaxRetry", "Max send retries").Prefix(prefix).DocPrefix("amqp").UintVar(fs, &config.MaxRetry, 3, overrides)
	flags.New("InactiveTimeout", "When inactive during the given timeout, stop listening").Prefix(prefix).DocPrefix("amqp").DurationVar(fs, &config.InactiveTimeout, 0, overrides)

	return &config
}

func New(config *Config, amqpClient *amqpclient.Client, metricProvider metric.MeterProvider, tracerProvider trace.TracerProvider, handler Handler) (*Service, error) {
	service := &Service{
		amqpClient:      amqpClient,
		exchange:        config.Exchange,
		queue:           config.Queue,
		exclusive:       config.Exclusive,
		routingKey:      config.RoutingKey,
		retryInterval:   config.RetryInterval,
		inactiveTimeout: config.InactiveTimeout,
		done:            make(chan struct{}),
		handler:         handler,
		maxRetry:        int64(config.MaxRetry),
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
		service.attributes = []attribute.KeyValue{
			semconv.MessagingSystemRabbitmq,
			semconv.NetworkProtocolName("rabbitmq"),
			semconv.MessagingDestinationName(service.exchange),
		}

		if len(service.routingKey) != 0 {
			service.attributes = append(service.attributes, semconv.MessagingRabbitmqDestinationRoutingKey(service.routingKey))
		}

		service.tracer = tracerProvider.Tracer("amqp_handler")
	}

	return service, nil
}

func (s *Service) Done() <-chan struct{} {
	return s.done
}

func (s *Service) Start(ctx context.Context) {
	defer close(s.done)

	if s.amqpClient == nil {
		return
	}

	init := true
	log := slog.With("exchange", s.exchange).With("queue", s.queue).With("routingKey", s.routingKey).With("vhost", s.amqpClient.Vhost())

	consumerName, messages, err := s.amqpClient.Listen(func() (string, error) {
		queueName, err := s.configure(init)
		init = false

		return queueName, err
	}, s.exchange, s.routingKey)
	if err != nil {
		log.ErrorContext(ctx, "listen", "error", err)

		return
	}

	log = log.With("name", consumerName)

	log.InfoContext(ctx, "Start listening messages")
	defer log.InfoContext(ctx, "End listening messages")

	var ticker *time.Ticker
	if s.inactiveTimeout != 0 {
		ticker = time.NewTicker(s.inactiveTimeout)
		defer ticker.Stop()

		tickerCtx, cancel := context.WithCancel(ctx)
		go func(ctx context.Context) {
			defer cancel()

			select {
			case <-ctx.Done():
			case <-ticker.C:
			}
		}(tickerCtx)

		ctx = tickerCtx
	}

	concurrent.ChanUntilDone(ctx, messages, func(message amqp.Delivery) {
		if ticker != nil {
			ticker.Reset(s.inactiveTimeout)
		}

		s.handleMessage(telemetry.ExtractContext(ctx, message.Headers), log, message)
	}, func() {
		if err := s.amqpClient.StopListener(consumerName); err != nil {
			log.ErrorContext(ctx, "stopping listener", "error", err)
		}
	})
}

func (s *Service) handleMessage(ctx context.Context, log *slog.Logger, message amqp.Delivery) {
	var err error

	ctx, end := telemetry.StartSpan(ctx, s.tracer, "receive",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			append([]attribute.KeyValue{
				semconv.MessagingOperationReceive,
			}, s.attributes...)...,
		),
	)
	defer end(&err)

	defer recoverer.Error(&err)

	err = s.handler(ctx, message)

	if err == nil {
		s.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "ack"),
			attribute.String("exchange", s.exchange),
			attribute.String("routingKey", s.routingKey),
		))
		if err = message.Ack(false); err != nil {
			log.ErrorContext(ctx, "ack message", "error", err)
		}

		return
	}

	log.ErrorContext(ctx, "handle message", "error", err, "body", string(message.Body))

	if s.retryInterval > 0 && s.maxRetry > 0 {
		s.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "retry"),
			attribute.String("exchange", s.exchange),
			attribute.String("routingKey", s.routingKey),
		))

		if err = s.Retry(message); err == nil {
			return
		}

		log.ErrorContext(ctx, "retry message", "error", err)
	}

	if err = message.Ack(false); err != nil {
		s.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "drop"),
			attribute.String("exchange", s.exchange),
			attribute.String("routingKey", s.routingKey),
		))

		log.ErrorContext(ctx, "ack message to trash it", "error", err)
	}
}

func (s *Service) configure(init bool) (string, error) {
	if !s.exclusive && !init {
		return s.queue, nil
	}

	queue := s.queue
	if s.exclusive {
		queue = fmt.Sprintf("%s-%s", s.queue, generateIdentityName())
	}

	if err := s.amqpClient.Consumer(queue, s.routingKey, s.exchange, s.exclusive, s.delayExchange); err != nil {
		return "", fmt.Errorf("configure amqp consumer for routingKey `%s` and exchange `%s`: %w", s.routingKey, s.exchange, err)
	}

	return queue, nil
}

func generateIdentityName() string {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "generate identity name", slog.Any("error", err))

		return "error"
	}

	return fmt.Sprintf("%x", raw)
}

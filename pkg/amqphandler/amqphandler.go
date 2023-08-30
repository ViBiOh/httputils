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

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Exchange", "Exchange name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Exchange, "", overrides)
	flags.New("Queue", "Queue name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.Queue, "", overrides)
	flags.New("Exclusive", "Queue exclusive mode (for fanout exchange)").Prefix(prefix).DocPrefix("amqp").BoolVar(fs, &config.Exclusive, false, overrides)
	flags.New("RoutingKey", "RoutingKey name").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.RoutingKey, "", overrides)
	flags.New("RetryInterval", "Interval duration when send fails").Prefix(prefix).DocPrefix("amqp").DurationVar(fs, &config.RetryInterval, time.Hour, overrides)
	flags.New("MaxRetry", "Max send retries").Prefix(prefix).DocPrefix("amqp").UintVar(fs, &config.MaxRetry, 3, overrides)

	return &config
}

func New(config *Config, amqpClient *amqpclient.Client, metricProvider metric.MeterProvider, tracerProvider trace.TracerProvider, handler Handler) (*Service, error) {
	service := &Service{
		amqpClient:    amqpClient,
		exchange:      config.Exchange,
		queue:         config.Queue,
		exclusive:     config.Exclusive,
		routingKey:    config.RoutingKey,
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
		log.Error("listen", "err", err)

		return
	}

	log = log.With("name", consumerName)

	go func() {
		<-ctx.Done()
		if err := s.amqpClient.StopListener(consumerName); err != nil {
			log.Error("error while stopping listener", "err", err)
		}
	}()

	log.Info("Start listening messages")
	defer log.Info("End listening messages")

	for message := range messages {
		s.handleMessage(ctx, log, message)
	}
}

func (s *Service) handleMessage(ctx context.Context, log *slog.Logger, message amqp.Delivery) {
	var err error

	ctx, end := telemetry.StartSpan(ctx, s.tracer, "handle", trace.WithSpanKind(trace.SpanKindConsumer))
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
			log.Error("ack message", "err", err)
		}

		return
	}

	log.Error("handle message", "err", err, "body", string(message.Body))

	if s.retryInterval > 0 && s.maxRetry > 0 {
		s.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "retry"),
			attribute.String("exchange", s.exchange),
			attribute.String("routingKey", s.routingKey),
		))

		if err = s.Retry(message); err == nil {
			return
		}

		log.Error("retry message", "err", err)
	}

	if err = message.Ack(false); err != nil {
		s.counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("state", "drop"),
			attribute.String("exchange", s.exchange),
			attribute.String("routingKey", s.routingKey),
		))

		log.Error("ack message to trash it", "err", err)
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
		slog.Error("generate identity name", "err", err)

		return "error"
	}

	return fmt.Sprintf("%x", raw)
}

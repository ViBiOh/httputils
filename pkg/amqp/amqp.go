package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source amqp.go -destination ../mocks/amqp.go -package mocks -mock_names Connection=AMQPConnection

var ErrNoConfig = errors.New("URI is required")

type Connection interface {
	io.Closer
	Channel() (*amqp.Channel, error)
	IsClosed() bool
}

type Client struct {
	tracer          trace.Tracer
	channel         *amqp.Channel
	connection      Connection
	listeners       map[string]*listener
	reconnectMetric metric.Int64Counter
	listenerMetric  metric.Int64UpDownCounter
	messageMetric   metric.Int64Counter
	vhost           string
	uri             string
	prefetch        int
	mutex           sync.RWMutex
}

type Config struct {
	URI      string
	Prefetch int
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("URI", "Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost>").Prefix(prefix).DocPrefix("amqp").StringVar(fs, &config.URI, "", overrides)
	flags.New("Prefetch", "Prefetch count for QoS").Prefix(prefix).DocPrefix("amqp").IntVar(fs, &config.Prefetch, 1, overrides)

	return &config
}

func New(config *Config, meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) (*Client, error) {
	return NewFromURI(config.URI, config.Prefetch, meterProvider, tracerProvider)
}

func NewFromURI(uri string, prefetch int, meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) (*Client, error) {
	if len(uri) == 0 {
		return nil, ErrNoConfig
	}

	client := &Client{
		uri:       uri,
		prefetch:  prefetch,
		listeners: make(map[string]*listener),
	}

	if meterProvider != nil {
		var err error

		client.reconnectMetric, client.listenerMetric, client.messageMetric, err = initMetrics(meterProvider)
		if err != nil {
			return nil, fmt.Errorf("init metrics: %w", err)
		}
	}

	if tracerProvider != nil {
		client.tracer = tracerProvider.Tracer("amqp")
	}

	connection, channel, err := connect(uri, client.prefetch, client.onDisconnect)
	if err != nil {
		return nil, fmt.Errorf("connect to amqp: %w", err)
	}

	client.connection = connection
	client.channel = channel
	client.vhost = connection.Config.Vhost

	slog.Info("Connected to AMQP!", "vhost", client.vhost)

	if err = client.Ping(); err != nil {
		return client, fmt.Errorf("ping amqp: %w", err)
	}

	return client, nil
}

func initMetrics(provider metric.MeterProvider) (metric.Int64Counter, metric.Int64UpDownCounter, metric.Int64Counter, error) {
	meter := provider.Meter("github.com/ViBiOh/httputils/v4/pkg/amqp")

	reconnect, err := meter.Int64Counter("amqp.reconnection")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create reconnection counter: %w", err)
	}

	listener, err := meter.Int64UpDownCounter("amqp.listener")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create listener counter: %w", err)
	}

	message, err := meter.Int64Counter("amqp.message")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create message counter: %w", err)
	}

	return reconnect, listener, message, nil
}

func (c *Client) Publish(ctx context.Context, payload amqp.Publishing, exchange, routingKey string) (err error) {
	if c == nil {
		return nil
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "publish", trace.WithSpanKind(trace.SpanKindProducer))
	defer end(&err)

	defer recoverer.Error(&err)

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if err = c.channel.PublishWithContext(ctx, exchange, routingKey, false, false, telemetry.InjectToAmqp(ctx, payload)); err != nil {
		c.increase(ctx, "error", exchange, routingKey)

		return
	}

	c.increase(ctx, "send", exchange, routingKey)

	return nil
}

func (c *Client) PublishJSON(ctx context.Context, item any, exchange, routingKey string) (err error) {
	if c == nil {
		return nil
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "publish_json", trace.WithSpanKind(trace.SpanKindProducer))
	defer end(&err)

	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = c.Publish(ctx, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	}, exchange, routingKey); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func (c *Client) increase(ctx context.Context, name, exchange, routingKey string) {
	if c.messageMetric == nil {
		return
	}

	c.messageMetric.Add(ctx, 1, metric.WithAttributes(
		attribute.String("state", name),
		attribute.String("exchange", exchange),
		attribute.String("routingKey", routingKey),
	))
}

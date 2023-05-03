package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/prometheus/client_golang/prometheus"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source amqp.go -destination ../mocks/amqp.go -package mocks -mock_names Connection=AMQPConnection

const (
	metricNamespace = "amqp"
)

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
	reconnectMetric prometheus.Counter
	listenerMetric  prometheus.Gauge
	messageMetrics  *prometheus.CounterVec
	vhost           string
	uri             string
	prefetch        int
	mutex           sync.RWMutex
}

type Config struct {
	uri      *string
	prefetch *int
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		uri:      flags.New("URI", "Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost>").Prefix(prefix).DocPrefix("amqp").String(fs, "", overrides),
		prefetch: flags.New("Prefetch", "Prefetch count for QoS").Prefix(prefix).DocPrefix("amqp").Int(fs, 1, overrides),
	}
}

func New(config Config, prometheusRegister prometheus.Registerer, tracer trace.Tracer) (*Client, error) {
	return NewFromURI(strings.TrimSpace(*config.uri), *config.prefetch, prometheusRegister, tracer)
}

func NewFromURI(uri string, prefetch int, prometheusRegister prometheus.Registerer, tracer trace.Tracer) (*Client, error) {
	if len(uri) == 0 {
		return nil, ErrNoConfig
	}

	client := &Client{
		tracer:          tracer,
		uri:             uri,
		prefetch:        prefetch,
		listeners:       make(map[string]*listener),
		reconnectMetric: prom.Counter(prometheusRegister, metricNamespace, "", "reconnection"),
		listenerMetric:  prom.Gauge(prometheusRegister, metricNamespace, "", "listener"),
		messageMetrics:  prom.CounterVec(prometheusRegister, metricNamespace, "", "message", "state", "exchange", "routingKey"),
	}

	connection, channel, err := connect(uri, client.prefetch, client.onDisconnect)
	if err != nil {
		return nil, fmt.Errorf("connect to amqp: %w", err)
	}

	client.connection = connection
	client.channel = channel
	client.vhost = connection.Config.Vhost

	logger.WithField("vhost", client.vhost).Info("Connected to AMQP!")

	if err = client.Ping(); err != nil {
		return client, fmt.Errorf("ping amqp: %w", err)
	}

	return client, nil
}

func (c *Client) Publish(ctx context.Context, payload amqp.Publishing, exchange, routingKey string) (err error) {
	_, end := tracer.StartSpan(ctx, c.tracer, "publish", trace.WithSpanKind(trace.SpanKindProducer))
	defer end(&err)

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if err = c.channel.PublishWithContext(ctx, exchange, routingKey, false, false, payload); err != nil {
		c.increase("error", exchange, routingKey)

		return
	}

	c.increase("published", exchange, routingKey)

	return nil
}

func (c *Client) PublishJSON(ctx context.Context, item any, exchange, routingKey string) (err error) {
	ctx, end := tracer.StartSpan(ctx, c.tracer, "publish_json", trace.WithSpanKind(trace.SpanKindProducer))
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

func (c *Client) increase(name, exchange, routingKey string) {
	if c.messageMetrics == nil {
		return
	}

	c.messageMetrics.WithLabelValues(name, exchange, routingKey).Inc()
}

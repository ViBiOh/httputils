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
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/ViBiOh/httputils/v4/pkg/waitcp"
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
	wait     *time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		uri:      flags.String(fs, prefix, "amqp", "URI", "Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost>", "", overrides),
		prefetch: flags.Int(fs, prefix, "amqp", "Prefetch", "Prefetch count for QoS", 1, overrides),
		wait:     flags.Duration(fs, prefix, "amqp", "WaitTimeout", "Wait duration for AMQP to be ready", time.Second*5, overrides),
	}
}

func New(config Config, prometheusRegister prometheus.Registerer, tracer trace.Tracer) (*Client, error) {
	return NewFromURI(strings.TrimSpace(*config.uri), *config.prefetch, *config.wait, prometheusRegister, tracer)
}

func NewFromURI(uri string, prefetch int, wait time.Duration, prometheusRegister prometheus.Registerer, tracer trace.Tracer) (*Client, error) {
	if len(uri) == 0 {
		return nil, ErrNoConfig
	}

	if wait > 0 {
		amqpURI, err := amqp.ParseURI(uri)
		if err != nil {
			return nil, fmt.Errorf("parse uri: %w", err)
		}

		if !waitcp.Wait("tcp", fmt.Sprintf("%s:%d", amqpURI.Host, amqpURI.Port), wait) {
			logger.Warn("database on `%s:%d` not ready", amqpURI.Host, amqpURI.Port)
		}
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

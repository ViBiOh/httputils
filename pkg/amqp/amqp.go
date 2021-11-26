package amqp

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	prom "github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/streadway/amqp"
)

const (
	metricNamespace = "amqp"
)

// ErrNoConfig occurs when URI is not provided
var ErrNoConfig = errors.New("URI is required")

// Connection for AMQP
//go:generate mockgen -destination ../mocks/amqp.go -mock_names Connection=AMQPConnection -package mocks github.com/ViBiOh/httputils/v4/pkg/amqp Connection
type Connection interface {
	io.Closer
	Channel() (*amqp.Channel, error)
	IsClosed() bool
}

// Client wraps all object required for AMQP usage
type Client struct {
	channel           *amqp.Channel
	connection        Connection
	listeners         map[string]*listener
	connectionMetrics map[string]prometheus.Counter
	messageMetrics    *prometheus.CounterVec
	vhost             string
	uri               string
	sync.RWMutex
}

// Config of package
type Config struct {
	uri *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		uri: flags.New(prefix, "amqp", "URI").Default("", nil).Label("Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost>").ToString(fs),
	}
}

// New inits AMQP connection from Config
func New(config Config, prometheusRegister prometheus.Registerer) (*Client, error) {
	return NewFromURI(strings.TrimSpace(*config.uri), prometheusRegister)
}

// NewFromURI inits AMQP connection from given URI
func NewFromURI(uri string, prometheusRegister prometheus.Registerer) (*Client, error) {
	if len(uri) == 0 {
		return nil, ErrNoConfig
	}

	client := &Client{
		uri:               uri,
		listeners:         make(map[string]*listener),
		connectionMetrics: prom.Counters(prometheusRegister, metricNamespace, "connection", "reconnect", "listener"),
		messageMetrics:    prom.CounterVec(prometheusRegister, metricNamespace, "", "message", "state"),
	}

	connection, channel, err := connect(uri, client.onDisconnect)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to amqp: %s", err)
	}

	client.connection = connection
	client.channel = channel
	client.vhost = connection.Config.Vhost

	logger.WithField("vhost", client.vhost).Info("Connected to AMQP!")

	if err = client.Ping(); err != nil {
		return client, fmt.Errorf("unable to ping amqp: %s", err)
	}

	return client, nil
}

// Publish sends payload to the underlying exchange
func (c *Client) Publish(payload amqp.Publishing, exchange, routingKey string) error {
	c.RLock()
	defer c.RUnlock()

	c.increase("published")

	return c.channel.Publish(exchange, routingKey, false, false, payload)
}

// PublishJSON sends JSON payload to the underlying exchange
func (c *Client) PublishJSON(item interface{}, exchange, routingKey string) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("unable to marshal: %s", err)
	}

	if err = c.Publish(amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	}, exchange, routingKey); err != nil {
		return fmt.Errorf("unable to publish: %s", err)
	}

	return nil
}

func (c *Client) increase(name string) {
	if c.messageMetrics == nil {
		return
	}

	c.messageMetrics.WithLabelValues(name).Inc()
}

func (c *Client) increaseConnection(name string) {
	if gauge, ok := c.connectionMetrics[name]; ok {
		gauge.Inc()
	}
}

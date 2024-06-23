package main

import (
	"flag"
	"os"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

type configuration struct {
	logger    *logger.Config
	alcotest  *alcotest.Config
	telemetry *telemetry.Config
	pprof     *pprof.Config
	health    *health.Config

	server   *server.Config
	owasp    *owasp.Config
	cors     *cors.Config
	renderer *renderer.Config

	amqp       *amqp.Config
	amqHandler *amqphandler.Config
	redis      *redis.Config
}

func newConfig() configuration {
	fs := flag.NewFlagSet("http", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:     logger.Flags(fs, "logger"),
		alcotest:   alcotest.Flags(fs, ""),
		telemetry:  telemetry.Flags(fs, "telemetry"),
		pprof:      pprof.Flags(fs, "pprof"),
		health:     health.Flags(fs, ""),
		server:     server.Flags(fs, ""),
		owasp:      owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'httputils-nonce'")),
		cors:       cors.Flags(fs, "cors"),
		amqp:       amqp.Flags(fs, "amqp"),
		amqHandler: amqphandler.Flags(fs, "amqp", flags.NewOverride("Exchange", "httputils"), flags.NewOverride("Queue", "httputils"), flags.NewOverride("RoutingKey", "local"), flags.NewOverride("RetryInterval", 10*time.Second)),
		redis:      redis.Flags(fs, "redis"),
		renderer:   renderer.Flags(fs, "renderer"),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}

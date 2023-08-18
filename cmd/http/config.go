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
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

type configuration struct {
	appServer  server.Config
	health     health.Config
	alcotest   alcotest.Config
	logger     logger.Config
	telemetry  telemetry.Config
	owasp      owasp.Config
	cors       cors.Config
	amqp       amqp.Config
	amqHandler amqphandler.Config
	redis      redis.Config
	renderer   renderer.Config
}

func newConfig() (configuration, error) {
	fs := flag.NewFlagSet("http", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	return configuration{
		appServer:  server.Flags(fs, ""),
		health:     health.Flags(fs, ""),
		alcotest:   alcotest.Flags(fs, ""),
		logger:     logger.Flags(fs, "logger"),
		telemetry:  telemetry.Flags(fs, "telemetry"),
		owasp:      owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'httputils-nonce'")),
		cors:       cors.Flags(fs, "cors"),
		amqp:       amqp.Flags(fs, "amqp"),
		amqHandler: amqphandler.Flags(fs, "amqp", flags.NewOverride("Exchange", "httputils"), flags.NewOverride("Queue", "httputils"), flags.NewOverride("RoutingKey", "local"), flags.NewOverride("RetryInterval", 10*time.Second)),
		redis:      redis.Flags(fs, "redis"),
		renderer:   renderer.Flags(fs, "renderer"),
	}, fs.Parse(os.Args[1:])
}

package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	amqplib "github.com/streadway/amqp"
)

//go:embed templates static
var content embed.FS

func main() {
	fs := flag.NewFlagSet("http", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	promServerConfig := server.Flags(fs, "prometheus", flags.NewOverride("Port", 9090), flags.NewOverride("IdleTimeout", "10s"), flags.NewOverride("ShutdownTimeout", "5s"))

	healthConfig := health.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'nonce'"))
	corsConfig := cors.Flags(fs, "cors")

	amqpConfig := amqp.Flags(fs, "amqp")
	amqHandlerConfig := amqphandler.Flags(fs, "amqp", flags.NewOverride("Exchange", "httputils"), flags.NewOverride("Queue", "httputils"), flags.NewOverride("RetryInterval", "10s"))

	rendererConfig := renderer.Flags(fs, "renderer")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	appServer := server.New(appServerConfig)
	promServer := server.New(promServerConfig)
	prometheusApp := prometheus.New(prometheusConfig)
	healthApp := health.New(healthConfig)

	amqpClient, err := amqp.New(amqpConfig, prometheusApp.Registerer())
	logger.Fatal(err)

	amqpApp, err := amqphandler.New(amqHandlerConfig, amqpClient, amqpHandler)
	logger.Fatal(err)

	logger.Fatal(amqpClient.Publisher("httputils", "direct", nil))

	rendererApp, err := renderer.New(rendererConfig, content, nil)
	logger.Fatal(err)

	speakingClock := cron.New().Each(5 * time.Minute).OnSignal(syscall.SIGUSR1).OnError(func(err error) {
		logger.Error("error while running cron: %s", err)
	}).Now()
	go speakingClock.Start(func(_ context.Context) error {
		logger.Info("Clock is ticking")
		return nil
	}, healthApp.Done())
	defer speakingClock.Shutdown()

	templateFunc := func(w http.ResponseWriter, r *http.Request) (string, int, map[string]interface{}, error) {
		return "public", http.StatusOK, nil, nil
	}

	go amqpApp.Start(healthApp.Done())
	go promServer.Start("prometheus", healthApp.End(), prometheusApp.Handler())
	go appServer.Start("http", healthApp.End(), httputils.Handler(rendererApp.Handler(templateFunc), healthApp, recoverer.Middleware, prometheusApp.Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), promServer.Done(), amqpApp.Done())
}

func amqpHandler(message amqplib.Delivery) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		return fmt.Errorf("unable to parse payload: %s", err)
	}

	return nil
}

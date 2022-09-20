package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	amqplib "github.com/streadway/amqp"
)

//go:embed templates static
var content embed.FS

func main() {
	config, err := newConfig()
	if err != nil {
		logger.Fatal(fmt.Errorf("config: %w", err))
	}

	alcotest.DoAndExit(config.alcotest)

	client, err := newClient(config)
	if err != nil {
		logger.Fatal(fmt.Errorf("client: %w", err))
	}

	adapter, err := newAdapter(config, client)
	if err != nil {
		logger.Fatal(fmt.Errorf("adapter: %w", err))
	}

	defer client.Close()

	appServer := server.New(config.appServer)
	promServer := server.New(config.promServer)

	ctx := client.health.Context()

	go client.redis.Pull(ctx, "httputils:tasks", func(content string, err error) {
		if err != nil {
			logger.Fatal(err)
		}

		logger.Info("content=`%s`", content)
	})

	speakingClock := cron.New().Each(5 * time.Minute).OnSignal(syscall.SIGUSR1).OnError(func(err error) {
		logger.Error("error while running cron: %s", err)
	}).Now()
	go speakingClock.Start(func(_ context.Context) error {
		logger.Info("Clock is ticking")

		return nil
	}, client.health.Done())
	defer speakingClock.Shutdown()

	templateFunc := func(w http.ResponseWriter, r *http.Request) (renderer.Page, error) {
		resp, err := request.Get("https://api.vibioh.fr/dump/").Send(r.Context(), nil)
		if err != nil {
			return renderer.Page{}, err
		}

		if err = request.DiscardBody(resp.Body); err != nil {
			return renderer.Page{}, err
		}

		return renderer.NewPage("public", http.StatusOK, nil), nil
	}

	go adapter.amqp.Start(context.Background(), client.health.Done())
	go promServer.Start("prometheus", client.health.End(), client.prometheus.Handler())
	go appServer.Start("http", client.health.End(), httputils.Handler(adapter.renderer.Handler(templateFunc), client.health, recoverer.Middleware, client.prometheus.Middleware, client.tracer.Middleware, owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	client.health.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), promServer.Done(), adapter.amqp.Done())
}

func amqpHandler(_ context.Context, message amqplib.Delivery) error {
	var payload map[string]any
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	return nil
}

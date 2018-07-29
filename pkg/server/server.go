package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/rollbar"
)

const healthcheckDuration = 35

func httpGracefulClose(server *http.Server) error {
	if server == nil {
		return nil
	}

	log.Print(`Shutting down HTTP server`)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf(`Error while shutting down HTTP server: %v`, err)
	}

	return nil
}

func gracefulClose(server *http.Server, callback func() error, healthcheckApp *healthcheck.App, flushers ...model.Flusher) int {
	exitCode := 0

	if healthcheckApp != nil {
		healthcheckApp.Close()
		log.Printf(`Waiting %d seconds for healthcheck`, healthcheckDuration)
		time.Sleep(time.Second * healthcheckDuration)
	}

	if err := httpGracefulClose(server); err != nil {
		rollbar.LogError(`%v`, err)
		exitCode = 1
	}

	if callback != nil {
		if err := callback(); err != nil {
			rollbar.LogError(`%v`, err)
			exitCode = 1
		}
	}

	for _, flusher := range flushers {
		flusher.Flush()
	}

	return exitCode
}

// GracefulClose gracefully close net/http server
func GracefulClose(server *http.Server, serveError <-chan error, callback func() error, healthcheckApp *healthcheck.App, flushers ...model.Flusher) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	select {
	case err := <-serveError:
		rollbar.LogError(`%v`, err)
	case <-signals:
		log.Print(`SIGTERM received`)
	}

	os.Exit(gracefulClose(server, callback, healthcheckApp, flushers...))
}

// ChainMiddlewares chains middlewares call for easy wrapping
func ChainMiddlewares(handler http.Handler, middlewares ...model.Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i].Handler(result)
	}

	return result
}

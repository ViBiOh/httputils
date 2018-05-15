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

func gracefulClose(server *http.Server, callback func() error, healthcheckApp *healthcheck.App) int {
	exitCode := 0

	if healthcheckApp != nil {
		healthcheckApp.Close()
		log.Printf(`Waiting %d seconds for healthcheck`, healthcheckDuration)
		time.Sleep(time.Second * healthcheckDuration)
	}

	if err := httpGracefulClose(server); err != nil {
		log.Print(err)
		exitCode = 1
	}

	if callback != nil {
		if err := callback(); err != nil {
			log.Print(err)
			exitCode = 1
		}
	}

	return exitCode
}

// GracefulClose gracefully close net/http server
func GracefulClose(server *http.Server, serveError <-chan error, callback func() error, healthcheckApp *healthcheck.App) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	select {
	case err := <-serveError:
		log.Print(err)
	case <-signals:
		log.Print(`SIGTERM received`)
	}

	os.Exit(gracefulClose(server, callback, healthcheckApp))
}

package httputils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func httpGracefulClose(server *http.Server) error {
	if server == nil {
		return nil
	}

	log.Print(`Shutting down http server`)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf(`Error while shutting down http server: %v`, err)
	}

	return nil
}

func gracefulClose(server *http.Server, callback func() error) int {
	exitCode := 0

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

// ServerGracefulClose gracefully close net/http server
func ServerGracefulClose(server *http.Server, callback func() error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	<-signals
	log.Printf(`SIGTERM received`)

	os.Exit(gracefulClose(server, callback))
}

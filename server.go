package httputils

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/httputils/cert"
)

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
func ServerGracefulClose(server *http.Server, serveError <-chan error, callback func() error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	select {
	case err := <-serveError:
		log.Print(err)
	case <-signals:
		log.Print(`SIGTERM received`)
	}

	os.Exit(gracefulClose(server, callback))
}

// StartMainServer starts server with common behavior and from a func that provide root handler
func StartMainServer(getHandler func() http.Handler, onGracefulClose func() error) {
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)

	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, false, `Serve TLS content`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting HTTP server on port %s`, *port)

	server := &http.Server{
		Addr:    fmt.Sprintf(`:%s`, *port),
		Handler: getHandler(),
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`👍 Listening with TLS`)
			serveError <- cert.ListenAndServeTLS(certConfig, server)
		} else {
			log.Print(`⚠ Listening without TLS`)
			serveError <- server.ListenAndServe()
		}
	}()

	ServerGracefulClose(server, serveError, onGracefulClose)
}

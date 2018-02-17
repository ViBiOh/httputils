package httputils

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/alcotest"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/server"
)

// StartMainServer starts server with common behavior and from a func that provide root handler
func StartMainServer(getHandler func() http.Handler, onGracefulClose func() error) {
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)

	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, false, `Serve TLS content`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting HTTP server on port %s`, *port)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(`:%s`, *port),
		Handler: getHandler(),
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`👍 Listening with TLS`)
			serveError <- cert.ListenAndServeTLS(certConfig, httpServer)
		} else {
			log.Print(`⚠ Listening without TLS`)
			serveError <- httpServer.ListenAndServe()
		}
	}()

	server.GracefulClose(httpServer, serveError, onGracefulClose)
}

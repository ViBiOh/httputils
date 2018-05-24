package httputils

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cert"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// App stores informations
type App struct {
	handlerFn       func() http.Handler
	gracefulCloseFn func() error
	healthcheckApp  *healthcheck.App
	port            *int
	tls             *bool
	alcotestConfig  map[string]*string
	certConfig      map[string]*string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, getHandler func() http.Handler, onGracefulClose func() error, healthcheckApp *healthcheck.App) *App {
	return &App{
		handlerFn:       getHandler,
		gracefulCloseFn: onGracefulClose,
		healthcheckApp:  healthcheckApp,
		port:            config[`port`].(*int),
		tls:             config[`tls`].(*bool),
		alcotestConfig:  config[`alcotestConfig`].(map[string]*string),
		certConfig:      config[`certConfig`].(map[string]*string),
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`port`:           flag.Int(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), 1080, `Listen port`),
		`tls`:            flag.Bool(tools.ToCamel(fmt.Sprintf(`%sTls`, prefix)), true, `Serve TLS content`),
		`certConfig`:     cert.Flags(`tls`),
		`alcotestConfig`: alcotest.Flags(``),
	}
}

// ListenAndServe starts server
func (a App) ListenAndServe() {
	flag.Parse()

	alcotest.DoAndExit(a.alcotestConfig)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(`:%d`, *a.port),
		Handler: a.handlerFn(),
	}

	log.Printf(`Starting HTTP server on port %s`, httpServer.Addr)

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *a.tls {
			log.Print(`👍 Listening with TLS`)
			serveError <- cert.ListenAndServeTLS(a.certConfig, httpServer)
		} else {
			log.Print(`⚠ Listening without TLS`)
			serveError <- httpServer.ListenAndServe()
		}
	}()

	server.GracefulClose(httpServer, serveError, a.gracefulCloseFn, a.healthcheckApp)
}

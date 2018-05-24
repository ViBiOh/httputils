package httputils

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/cert"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App stores informations
type App struct {
	port       *int
	tls        *bool
	certConfig map[string]*string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) *App {
	return &App{
		port:       config[`port`].(*int),
		tls:        config[`tls`].(*bool),
		certConfig: config[`certConfig`].(map[string]*string),
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`port`:       flag.Int(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), 1080, `Listen port`),
		`tls`:        flag.Bool(tools.ToCamel(fmt.Sprintf(`%sTls`, prefix)), true, `Serve TLS content`),
		`certConfig`: cert.Flags(`tls`),
	}
}

// ListenAndServe starts server
func (a App) ListenAndServe(handler http.Handler, onGracefulClose func() error, healthcheckApp *healthcheck.App) {
	healthcheckHandler := healthcheckApp.Handler()
	prometheusHandler := promhttp.Handler()

	httpServer := &http.Server{
		Addr: fmt.Sprintf(`:%d`, *a.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case `/health`:
				healthcheckHandler.ServeHTTP(w, r)
			case `/metrics`:
				prometheusHandler.ServeHTTP(w, r)
			default:
				handler.ServeHTTP(w, r)
			}
		}),
	}

	log.Printf(`Starting HTTP server on port %s`, httpServer.Addr)

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *a.tls {
			log.Print(`Listening with TLS ✅`)
			serveError <- cert.ListenAndServeTLS(a.certConfig, httpServer)
		} else {
			log.Print(`Listening without TLS ⚠️`)
			serveError <- httpServer.ListenAndServe()
		}
	}()

	server.GracefulClose(httpServer, serveError, onGracefulClose, healthcheckApp)
}

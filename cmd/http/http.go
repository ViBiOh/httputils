package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
)

func main() {
	fs := flag.NewFlagSet("http", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	httputils.New(serverConfig).ListenAndServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("It works!")); err != nil {
			logger.Warn("unable to write: %s", err)
		}
	}), nil, prometheus.New(prometheusConfig).Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware)
}

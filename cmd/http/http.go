package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
)

func main() {
	fs := flag.NewFlagSet("http", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")
	swaggerConfig := swagger.Flags(fs, "swagger")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))

	prometheusApp := prometheus.New(prometheusConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheusApp.Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger, prometheusApp.Swagger)
	logger.Fatal(err)

	server.ListenServeWait(swaggerApp.Handler())
}

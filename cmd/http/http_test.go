package main

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func benchmarkHandler(b *testing.B, handler http.Handler) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for b.Loop() {
		handler.ServeHTTP(writer, request)
	}
}

func BenchmarkNoMiddleware(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkFullMiddlewares", flag.ContinueOnError)

	fs.String("test.timeout", "", "")
	fs.String("test.bench", "", "")
	fs.String("test.benchmem", "", "")
	fs.String("test.run", "", "")
	fs.String("test.paniconexit0", "", "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		b.Error(err)
	}

	benchmarkHandler(b, handler)
}

func BenchmarkFullMiddlewares(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkFullMiddlewares", flag.ContinueOnError)

	fs.String("test.timeout", "", "")
	fs.String("test.bench", "", "")
	fs.String("test.benchmem", "", "")
	fs.String("test.run", "", "")
	fs.String("test.paniconexit0", "", "")

	telemetryConfig := telemetry.Flags(fs, "telemetry")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	if err := fs.Parse(os.Args[1:]); err != nil {
		b.Error(err)
	}

	telemetryService, err := telemetry.New(context.Background(), telemetryConfig)
	if err != nil {
		b.Error(err)
	}

	middlewares := model.ChainMiddlewares(handler, telemetryService.Middleware("http"), owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware)
	benchmarkHandler(b, middlewares)
}

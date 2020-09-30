package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
)

func benchmarkHandler(b *testing.B, handler http.Handler) {
	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	defaultHTTPClient := http.Client{
		Timeout: 30 * time.Second,
	}
	request := httptest.NewRequest(http.MethodGet, testServer.URL+"/", nil)

	for i := 0; i < b.N; i++ {
		defaultHTTPClient.Do(request)
	}
}

func BenchmarkNoMiddleware(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkFullMiddlewares", flag.ContinueOnError)

	fs.String("test.timeout", "", "")
	fs.String("test.bench", "", "")
	fs.String("test.benchmem", "", "")
	fs.String("test.run", "", "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		b.Error(err)
	}

	benchmarkHandler(b, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func BenchmarkFullMiddlewares(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkFullMiddlewares", flag.ContinueOnError)

	fs.String("test.timeout", "", "")
	fs.String("test.bench", "", "")
	fs.String("test.benchmem", "", "")
	fs.String("test.run", "", "")

	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	if err := fs.Parse(os.Args[1:]); err != nil {
		b.Error(err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler = httputils.ChainMiddlewares(handler, prometheus.New(prometheusConfig).Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware)
	benchmarkHandler(b, handler)
}

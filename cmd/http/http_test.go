package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/prometheus"
)

func benchmarkHandler(b *testing.B, handler http.Handler) {
	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	defaultHTTPClient := http.Client{
		Timeout: 30 * time.Second,
	}
	request, err := http.NewRequest(http.MethodGet, testServer.URL+"/", nil)
	if err != nil {
		b.Errorf("unable to create request: %s", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := defaultHTTPClient.Do(request); err != nil {
			b.Errorf("unable to execute request: %s", err)
		}
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

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	benchmarkHandler(b, handler)
}

func BenchmarkFullMiddlewares(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkFullMiddlewares", flag.ContinueOnError)

	fs.String("test.timeout", "", "")
	fs.String("test.bench", "", "")
	fs.String("test.benchmem", "", "")
	fs.String("test.run", "", "")
	fs.String("test.paniconexit0", "", "")

	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	if err := fs.Parse(os.Args[1:]); err != nil {
		b.Error(err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler = prometheus.New(prometheusConfig).Middleware(owasp.New(owaspConfig).Middleware(cors.New(corsConfig).Middleware(handler)))
	benchmarkHandler(b, handler)
}

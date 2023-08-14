package logger

import (
	"io"
	"log"
	"log/slog"
	"testing"
	"time"
)

func BenchmarkStandardSimpleOutput(b *testing.B) {
	logger := log.New(io.Discard, "", log.Ldate|log.Ltime)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print("[INFO] Hello world")
	}
}

func BenchmarkStandardSimpleFormattedOutput(b *testing.B) {
	logger := log.New(io.Discard, "", log.Ldate|log.Ltime)
	now := time.Now().Unix()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Printf("Hello %s, it's %d", "Bob", now)
	}
}

func BenchmarkStructuredNoOutput(b *testing.B) {
	configureLogger(io.Discard, slog.LevelError, false, "time", "level", "msg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slog.Info("Hello", "name", "Bob")
	}
}

func BenchmarkStructuredJSONOutput(b *testing.B) {
	configureLogger(io.Discard, slog.LevelInfo, true, "time", "level", "msg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slog.Info("Hello")
	}
}

func BenchmarkStructuredJSONOutputWithFields(b *testing.B) {
	configureLogger(io.Discard, slog.LevelInfo, true, "time", "level", "msg")

	logger := slog.
		With(slog.Int("count", 7)).
		With(slog.String("name", "test")).
		With(slog.Bool("success", true))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Hello")
	}
}

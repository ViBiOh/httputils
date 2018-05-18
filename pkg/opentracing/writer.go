package opentracing

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

// ResponseWriter implements http.ResponseWriter by storing content in memory
type ResponseWriter struct {
	http.ResponseWriter
	opentracing.Span
	status int
}

// NewResponseWriter creates a new response writer
func NewResponseWriter(w http.ResponseWriter, span opentracing.Span) *ResponseWriter {
	return &ResponseWriter{w, span, 0}
}

// WriteHeader cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *ResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)

	w.status = status
	w.SetTag(`http.status_code`, status)
	if status >= http.StatusBadRequest {
		w.SetTag(`error`, true)
	}
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

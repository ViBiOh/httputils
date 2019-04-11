package owasp

import (
	"net/http"

	"github.com/ViBiOh/httputils/pkg/model"
)

const (
	cacheControlHeader = "Cache-Control"
)

var _ model.Middleware = &App{}

type middleware struct {
	http.ResponseWriter
	index       bool
	wroteHeader bool
}

func (m *middleware) setHeader() {
	if m.Header().Get(cacheControlHeader) == "" {
		if m.index {
			m.Header().Set(cacheControlHeader, "no-cache")
		} else {
			m.Header().Set(cacheControlHeader, "max-age=864000")
		}
	}
}

func (m *middleware) WriteHeader(status int) {
	m.wroteHeader = true

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		m.setHeader()
	}

	m.ResponseWriter.WriteHeader(status)
}

func (m *middleware) Write(b []byte) (int, error) {
	if !m.wroteHeader {
		m.setHeader()
	}

	return m.ResponseWriter.Write(b)
}

func (m *middleware) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := m.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (m *middleware) Flush() {
	if f, ok := m.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

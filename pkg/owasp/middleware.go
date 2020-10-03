package owasp

import (
	"net/http"
)

const (
	cacheControlHeader = "Cache-Control"
)

type middleware struct {
	http.ResponseWriter
	index      bool
	cacheAdded bool
}

func (m *middleware) setCacheControl(status int) {
	m.cacheAdded = true

	if len(m.Header().Get(cacheControlHeader)) != 0 {
		return
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently || status == http.StatusSeeOther || status == http.StatusNotModified || status == http.StatusPermanentRedirect {
		if m.index {
			m.Header().Set(cacheControlHeader, "no-cache")
		} else {
			m.Header().Set(cacheControlHeader, "public, max-age=864000")
		}
	}
}

func (m *middleware) WriteHeader(status int) {
	m.setCacheControl(status)
	m.ResponseWriter.WriteHeader(status)
}

func (m *middleware) Write(b []byte) (int, error) {
	if !m.cacheAdded {
		m.setCacheControl(http.StatusOK)
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

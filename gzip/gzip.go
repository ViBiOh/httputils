package gzip

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var ignoreFile = regexp.MustCompile(`.png$`)
var acceptGzip = regexp.MustCompile(`^(?:gzip|\*)(?:;q=(?:1.*?|0\.[1-9][0-9]*))?$`)

type gzipMiddleware struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

func (m *gzipMiddleware) WriteHeader(status int) {
	m.ResponseWriter.Header().Add(`Vary`, `Accept-Encoding`)
	m.ResponseWriter.Header().Set(`Content-Encoding`, `gzip`)
	m.ResponseWriter.Header().Del(`Content-Length`)
}

func (m *gzipMiddleware) Write(b []byte) (int, error) {
	return m.gzw.Write(b)
}

func (m *gzipMiddleware) Flush() {
	m.gzw.Flush()

	if flusher, ok := m.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *gzipMiddleware) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := m.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf(`http.Hijacker not available`)
}

type gzipHandler struct {
	h http.Handler
}

func (handler gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if acceptEncodingGzip(r) && !ignoreFile.MatchString(r.URL.Path) {
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()

		handler.h.ServeHTTP(&gzipMiddleware{w, gzipWriter}, r)
	} else {
		handler.h.ServeHTTP(w, r)
	}
}

func acceptEncodingGzip(r *http.Request) bool {
	header := r.Header.Get(`Accept-Encoding`)

	for _, headerEncoding := range strings.Split(header, `,`) {
		if acceptGzip.MatchString(headerEncoding) {
			return true
		}
	}

	return false
}

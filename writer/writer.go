package writer

import (
	"bytes"
	"net/http"
)

// ResponseWriter implements http.ResponseWriter by storing content in memory
type ResponseWriter struct {
	status  int
	header  http.Header
	content *bytes.Buffer
}

// Content return current contant buffer
func (w *ResponseWriter) Content() *bytes.Buffer {
	return w.content
}

// Status return current writer status
func (w *ResponseWriter) Status() int {
	return w.status
}

// SetStatus set current writer status
func (w *ResponseWriter) SetStatus(status int) {
	w.status = status
}

// Header cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *ResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}

	return w.header
}

// Write cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *ResponseWriter) Write(content []byte) (int, error) {
	if w.content == nil {
		w.content = bytes.NewBuffer(make([]byte, 0, 1024))
	}

	return w.content.Write(content)
}

// WriteHeader cf. https://golang.org/pkg/net/http/#ResponseWriter
func (w *ResponseWriter) WriteHeader(status int) {
	w.status = status
}

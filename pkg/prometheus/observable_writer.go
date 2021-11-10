package prometheus

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	flusher = 1 << iota
	hijacker
	readerFrom
	pusher
)

var writers = make([]func(*observableResponseWriter) responseWriter, 16)

type responseWriter interface {
	http.ResponseWriter
	Status() int
	Written() int64
}

type observableResponseWriter struct {
	http.ResponseWriter

	status  int
	written int64
}

func (r *observableResponseWriter) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}

	return r.status
}

func (r *observableResponseWriter) Written() int64 {
	return r.written
}

func (r *observableResponseWriter) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *observableResponseWriter) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func newDelegator(w http.ResponseWriter) responseWriter {
	d := &observableResponseWriter{
		ResponseWriter: w,
	}

	id := 0
	if _, ok := w.(http.Flusher); ok {
		id |= flusher
	}
	if _, ok := w.(http.Hijacker); ok {
		id |= hijacker
	}
	if _, ok := w.(io.ReaderFrom); ok {
		id |= readerFrom
	}
	if _, ok := w.(http.Pusher); ok {
		id |= pusher
	}

	return writers[id](d)
}

type flusherDelegator struct{ *observableResponseWriter }

func (d flusherDelegator) Flush() {
	d.ResponseWriter.(http.Flusher).Flush()
}

type hijackerDelegator struct{ *observableResponseWriter }

func (d hijackerDelegator) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return d.ResponseWriter.(http.Hijacker).Hijack()
}

type readerFromDelegator struct{ *observableResponseWriter }

func (d readerFromDelegator) ReadFrom(r io.Reader) (int64, error) {
	n, err := d.ResponseWriter.(io.ReaderFrom).ReadFrom(r)
	d.written += n
	return n, err
}

type pusherDelegator struct{ *observableResponseWriter }

func (d pusherDelegator) Push(target string, opts *http.PushOptions) error {
	return d.ResponseWriter.(http.Pusher).Push(target, opts)
}

func init() {
	writers[0] = func(d *observableResponseWriter) responseWriter {
		return d
	}

	writers[flusher] = func(d *observableResponseWriter) responseWriter {
		return flusherDelegator{d}
	}

	writers[hijacker] = func(d *observableResponseWriter) responseWriter {
		return hijackerDelegator{d}
	}

	writers[hijacker|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Hijacker
			http.Flusher
		}{d, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[readerFrom] = func(d *observableResponseWriter) responseWriter {
		return readerFromDelegator{d}
	}

	writers[readerFrom|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			io.ReaderFrom
			http.Flusher
		}{d, readerFromDelegator{d}, flusherDelegator{d}}
	}

	writers[readerFrom|hijacker] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			io.ReaderFrom
			http.Hijacker
		}{d, readerFromDelegator{d}, hijackerDelegator{d}}
	}

	writers[readerFrom|hijacker|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			io.ReaderFrom
			http.Hijacker
			http.Flusher
		}{d, readerFromDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher] = func(d *observableResponseWriter) responseWriter {
		return pusherDelegator{d}
	}

	writers[pusher|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			http.Flusher
		}{d, pusherDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|hijacker] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			http.Hijacker
		}{d, pusherDelegator{d}, hijackerDelegator{d}}
	}

	writers[pusher|hijacker|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			http.Hijacker
			http.Flusher
		}{d, pusherDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|readerFrom] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			io.ReaderFrom
		}{d, pusherDelegator{d}, readerFromDelegator{d}}
	}

	writers[pusher|readerFrom|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			io.ReaderFrom
			http.Flusher
		}{d, pusherDelegator{d}, readerFromDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|readerFrom|hijacker] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			io.ReaderFrom
			http.Hijacker
		}{d, pusherDelegator{d}, readerFromDelegator{d}, hijackerDelegator{d}}
	}

	writers[pusher|readerFrom|hijacker|flusher] = func(d *observableResponseWriter) responseWriter {
		return struct {
			*observableResponseWriter
			http.Pusher
			io.ReaderFrom
			http.Hijacker
			http.Flusher
		}{d, pusherDelegator{d}, readerFromDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}
}

package owasp

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	flusher = 1 << iota
	hijacker
	readerFrom
	pusher
)

var writers = make([]func(*nonceResponseWritter) responseWriter, 16)

type responseWriter interface {
	http.ResponseWriter
	WriteNonce(string)
}

type nonceResponseWritter struct {
	http.ResponseWriter

	csp     string
	written bool
}

func (r *nonceResponseWritter) WriteNonce(nonce string) {
	r.Header().Add(cspHeader, strings.ReplaceAll(r.csp, nonceKey, "nonce-"+nonce))
	r.written = true
}

func (r *nonceResponseWritter) WriteHeader(code int) {
	if !r.written {
		r.WriteNonce(Nonce())
	}

	r.ResponseWriter.WriteHeader(code)
}

func (r *nonceResponseWritter) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteNonce(Nonce())
	}

	return r.ResponseWriter.Write(b)
}

func newDelegator(w http.ResponseWriter) responseWriter {
	d := &nonceResponseWritter{
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

type flusherDelegator struct{ *nonceResponseWritter }

func (d flusherDelegator) Flush() {
	d.ResponseWriter.(http.Flusher).Flush()
}

type hijackerDelegator struct{ *nonceResponseWritter }

func (d hijackerDelegator) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return d.ResponseWriter.(http.Hijacker).Hijack()
}

type readerFromDelegator struct{ *nonceResponseWritter }

func (d readerFromDelegator) ReadFrom(r io.Reader) (int64, error) {
	return d.ResponseWriter.(io.ReaderFrom).ReadFrom(r)
}

type pusherDelegator struct{ *nonceResponseWritter }

func (d pusherDelegator) Push(target string, opts *http.PushOptions) error {
	return d.ResponseWriter.(http.Pusher).Push(target, opts)
}

func init() {
	writers[0] = func(d *nonceResponseWritter) responseWriter {
		return d
	}

	writers[flusher] = func(d *nonceResponseWritter) responseWriter {
		return flusherDelegator{d}
	}

	writers[hijacker] = func(d *nonceResponseWritter) responseWriter {
		return hijackerDelegator{d}
	}

	writers[hijacker|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Hijacker
			http.Flusher
		}{d, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[readerFrom] = func(d *nonceResponseWritter) responseWriter {
		return readerFromDelegator{d}
	}

	writers[readerFrom|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			io.ReaderFrom
			http.Flusher
		}{d, readerFromDelegator{d}, flusherDelegator{d}}
	}

	writers[readerFrom|hijacker] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			io.ReaderFrom
			http.Hijacker
		}{d, readerFromDelegator{d}, hijackerDelegator{d}}
	}

	writers[readerFrom|hijacker|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			io.ReaderFrom
			http.Hijacker
			http.Flusher
		}{d, readerFromDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher] = func(d *nonceResponseWritter) responseWriter {
		return pusherDelegator{d}
	}

	writers[pusher|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			http.Flusher
		}{d, pusherDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|hijacker] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			http.Hijacker
		}{d, pusherDelegator{d}, hijackerDelegator{d}}
	}

	writers[pusher|hijacker|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			http.Hijacker
			http.Flusher
		}{d, pusherDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|readerFrom] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			io.ReaderFrom
		}{d, pusherDelegator{d}, readerFromDelegator{d}}
	}

	writers[pusher|readerFrom|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			io.ReaderFrom
			http.Flusher
		}{d, pusherDelegator{d}, readerFromDelegator{d}, flusherDelegator{d}}
	}

	writers[pusher|readerFrom|hijacker] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			io.ReaderFrom
			http.Hijacker
		}{d, pusherDelegator{d}, readerFromDelegator{d}, hijackerDelegator{d}}
	}

	writers[pusher|readerFrom|hijacker|flusher] = func(d *nonceResponseWritter) responseWriter {
		return struct {
			*nonceResponseWritter
			http.Pusher
			io.ReaderFrom
			http.Hijacker
			http.Flusher
		}{d, pusherDelegator{d}, readerFromDelegator{d}, hijackerDelegator{d}, flusherDelegator{d}}
	}
}

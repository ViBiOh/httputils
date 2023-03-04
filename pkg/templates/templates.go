package templates

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
	"go.opentelemetry.io/otel/trace"
)

var (
	minifier *minify.M

	htmlHeaders = http.Header{}
	xmlHeaders  = http.Header{}

	bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 4*1024))
		},
	}
)

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	minifier.AddFunc("text/xml", svg.Minify)

	htmlHeaders.Add("X-UA-Compatible", "ie=edge")
	htmlHeaders.Add("Content-Type", "text/html; charset=UTF-8")
	htmlHeaders.Add("Cache-Control", "no-cache")

	xmlHeaders.Add("Content-Type", "text/xml; charset=UTF-8")
	xmlHeaders.Add("Cache-Control", "no-cache")
}

func minifyWithTracing(ctx context.Context, tr trace.Tracer, mediatype string, input io.Reader, output io.Writer) (err error) {
	_, end := tracer.StartSpan(ctx, tr, "minify", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	return minifier.Minify(mediatype, output, input)
}

func WriteTemplate(ctx context.Context, tr trace.Tracer, tpl *template.Template, w io.Writer, content any, mediatype string) (err error) {
	_, end := tracer.StartSpan(ctx, tr, "template", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	if err := tpl.Execute(buffer, content); err != nil {
		return err
	}

	return minifyWithTracing(ctx, tr, mediatype, buffer, w)
}

func ResponseHTMLTemplate(ctx context.Context, tr trace.Tracer, tpl *template.Template, w http.ResponseWriter, content any, status int) (err error) {
	ctx, end := tracer.StartSpan(ctx, tr, "html_template", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	if err := tpl.Execute(buffer, content); err != nil {
		return err
	}

	for key, value := range htmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return minifyWithTracing(ctx, tr, "text/html", buffer, w)
}

func ResponseHTMLTemplateRaw(ctx context.Context, tr trace.Tracer, tpl *template.Template, w http.ResponseWriter, content any, status int) (err error) {
	_, end := tracer.StartSpan(ctx, tr, "html_template_raw", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	for key, value := range htmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return tpl.Execute(w, content)
}

func ResponseXMLTemplate(ctx context.Context, tr trace.Tracer, tpl *template.Template, w http.ResponseWriter, content any, status int) (err error) {
	ctx, end := tracer.StartSpan(ctx, tr, "xml_template", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	buffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	if err := tpl.Execute(buffer, content); err != nil {
		return err
	}

	for key, value := range xmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return minifyWithTracing(ctx, tr, "text/xml", buffer, w)
}

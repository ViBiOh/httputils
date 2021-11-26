package templates

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

var (
	minifier *minify.M

	htmlHeaders = http.Header{}
	xmlHeaders  = http.Header{}

	bufferPool = sync.Pool{
		New: func() interface{} {
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

// WriteTemplateRaw write template name from given template into writer for provided content
func WriteTemplateRaw(tpl *template.Template, w io.Writer, content interface{}) error {
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	return tpl.Execute(buffer, content)
}

// WriteTemplate write template name from given template into writer for provided content with given minification
func WriteTemplate(tpl *template.Template, w io.Writer, content interface{}, mediatype string) error {
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	if err := tpl.Execute(buffer, content); err != nil {
		return err
	}

	return minifier.Minify(mediatype, w, buffer)
}

// ResponseHTMLTemplate write template name from given template into writer for provided content with HTML minification
func ResponseHTMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
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

	return minifier.Minify("text/html", w, buffer)
}

// ResponseHTMLTemplateRaw write template name from given template into writer for provided content
func ResponseHTMLTemplateRaw(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	for key, value := range htmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return tpl.Execute(w, content)
}

// ResponseXMLTemplate write template name from given template into writer for provided content with XML minification
func ResponseXMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
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

	return minifier.Minify("text/xml", w, buffer)
}

package templates

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"regexp"

	"github.com/ViBiOh/httputils/v4/pkg/model"
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

// WriteTemplate write template name from given template into writer for provided content with given minification
func WriteTemplate(tpl *template.Template, w io.Writer, content interface{}, mediatype string) error {
	templateBuffer := model.BufferPool.Get().(*bytes.Buffer)
	defer model.BufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	return minifier.Minify(mediatype, w, templateBuffer)
}

// ResponseHTMLTemplate write template name from given template into writer for provided content with HTML minification
func ResponseHTMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := model.BufferPool.Get().(*bytes.Buffer)
	defer model.BufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	for key, value := range htmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return minifier.Minify("text/html", w, templateBuffer)
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
	templateBuffer := model.BufferPool.Get().(*bytes.Buffer)
	defer model.BufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	templateBuffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	for key, value := range xmlHeaders {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	return minifier.Minify("text/xml", w, templateBuffer)
}

package httputils

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/xml"
)

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc(`text/html`, html.Minify)
	minifier.AddFunc(`text/css`, css.Minify)
	minifier.AddFunc(`text/javascript`, js.Minify)
	minifier.AddFunc(`text/xml`, xml.Minify)
}

// WriteHTMLTemplate write template name from given template into writer for provided content with HTML minification
func WriteHTMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := &bytes.Buffer{}
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	w.WriteHeader(status)
	w.Header().Add(`Content-Type`, `text/html; charset=UTF-8`)
	w.Header().Add(`Cache-Control`, `no-cache`)
	w.Header().Add(`X-UA-Compatible`, `IE=edge,chrome=1`)
	minifier.Minify(`text/html`, w, templateBuffer)
	return nil
}

// WriteXMLTemplate write template name from given template into writer for provided content with XML minification
func WriteXMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := &bytes.Buffer{}
	templateBuffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	w.WriteHeader(status)
	w.Header().Add(`Content-Type`, `text/xml; charset=UTF-8`)
	w.Header().Add(`Cache-Control`, `no-cache`)
	minifier.Minify(`text/xml`, w, templateBuffer)
	return nil
}

package templates

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
)

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	minifier.AddFunc("text/xml", svg.Minify)
}

// GetTemplates list files by extension
func GetTemplates(dir, ext string) ([]string, error) {
	var output []string

	if err := filepath.Walk(dir, func(walkedPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.Ext(info.Name()) == ext {
			output = append(output, walkedPath)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return output, nil
}

// WriteTemplate write template name from given template into writer for provided content with given minification
func WriteTemplate(tpl *template.Template, w io.Writer, content interface{}, mediatype string) error {
	templateBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	return minifier.Minify(mediatype, w, templateBuffer)
}

// ResponseHTMLTemplate write template name from given template into writer for provided content with HTML minification
func ResponseHTMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	w.Header().Add("X-UA-Compatible", "ie=edge")
	contentType(w, "text/html; charset=UTF-8")
	noCache(w)
	w.WriteHeader(status)
	return minifier.Minify("text/html", w, templateBuffer)
}

// ResponseHTMLTemplateRaw write template name from given template into writer for provided content
func ResponseHTMLTemplateRaw(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	w.Header().Add("X-UA-Compatible", "ie=edge")
	contentType(w, "text/html; charset=UTF-8")
	noCache(w)
	w.WriteHeader(status)

	return tpl.Execute(w, content)
}

// ResponseXMLTemplate write template name from given template into writer for provided content with XML minification
func ResponseXMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(templateBuffer)

	templateBuffer.Reset()
	templateBuffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return err
	}

	contentType(w, "text/xml; charset=UTF-8")
	noCache(w)
	w.WriteHeader(status)
	return minifier.Minify("text/xml", w, templateBuffer)
}

func contentType(w http.ResponseWriter, contentType string) {
	w.Header().Add("Content-Type", contentType)
}

func noCache(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-cache")
}

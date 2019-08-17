package templates

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/svg"
)

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/xml", svg.Minify)
}

// GetTemplates list files by extension
func GetTemplates(dir, ext string) ([]string, error) {
	output := make([]string, 0)

	if err := filepath.Walk(dir, func(walkedPath string, info os.FileInfo, _ error) error {
		if path.Ext(info.Name()) == ext {
			output = append(output, walkedPath)
		}

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return output, nil
}

// WriteHTMLTemplate write template name from given template into writer for provided content with HTML minification
func WriteHTMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := &bytes.Buffer{}
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return errors.WithStack(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-UA-Compatible", "ie=edge")
	w.WriteHeader(status)
	return errors.WithStack(minifier.Minify("text/html", w, templateBuffer))
}

// WriteXMLTemplate write template name from given template into writer for provided content with XML minification
func WriteXMLTemplate(tpl *template.Template, w http.ResponseWriter, content interface{}, status int) error {
	templateBuffer := &bytes.Buffer{}
	templateBuffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	if err := tpl.Execute(templateBuffer, content); err != nil {
		return errors.WithStack(err)
	}

	w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	return errors.WithStack(minifier.Minify("text/xml", w, templateBuffer))
}

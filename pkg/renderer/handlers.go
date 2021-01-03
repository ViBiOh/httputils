package renderer

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/renderer/model"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

// Redirect redirect user to a defined path with a message
func Redirect(w http.ResponseWriter, r *http.Request, pathname string, message model.Message) {
	var redirect string

	value, err := url.Parse(pathname)
	if err == nil && value.Host != "" {
		redirect = pathname
	} else {
		redirect = fmt.Sprintf("%s://%s", r.URL.Scheme, path.Join(r.URL.Host, pathname))
	}

	http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect, message), http.StatusFound)
}

func (a app) Error(w http.ResponseWriter, err error) {
	logger.Error("%s", err)
	content := a.feedContent(nil)

	status, message := model.ErrorStatus(err)
	if len(message) > 0 {
		content["Message"] = model.NewErrorMessage(message)
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) html(w http.ResponseWriter, r *http.Request, templateFunc model.TemplateFunc) {
	templateName, status, content, err := templateFunc(r)
	if err != nil {
		a.Error(w, err)
		return
	}

	content = a.feedContent(content)

	message := model.ParseMessage(r)
	if len(message.Content) > 0 {
		content["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup(templateName), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) svg() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup(fmt.Sprintf("svg-%s", strings.Trim(r.URL.Path, "/")))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := templates.WriteTemplate(tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

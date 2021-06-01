package renderer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
)

// Redirect redirect user to a defined path with a message
func (a app) Redirect(w http.ResponseWriter, r *http.Request, pathname string, message Message) {
	joinChar := "?"
	if strings.Contains(pathname, "?") {
		joinChar = "&"
	}

	var anchor string
	parts := strings.SplitN(pathname, "#", 2)
	if len(parts) == 2 {
		anchor = fmt.Sprintf("#%s", parts[1])
	}

	http.Redirect(w, r, fmt.Sprintf("%s%s%s%s", a.url(parts[0]), joinChar, message, anchor), http.StatusFound)
}

func (a app) Error(w http.ResponseWriter, err error) {
	logger.Error("%s", err)
	content := a.feedContent(nil)

	status, message := httperror.ErrorStatus(err)
	if len(message) > 0 {
		content["Message"] = NewErrorMessage(message)
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) render(w http.ResponseWriter, r *http.Request, templateFunc TemplateFunc) {
	defer func() {
		if r := recover(); r != nil {
			a.Error(w, fmt.Errorf("recovered from panic: %s", r))
		}
	}()

	templateName, status, content, err := templateFunc(w, r)
	if err != nil {
		a.Error(w, err)
		return
	}

	if len(templateName) == 0 {
		return
	}

	content = a.feedContent(content)

	message := ParseMessage(r)
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

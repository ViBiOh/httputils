package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/renderer/model"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

func (a app) error(w http.ResponseWriter, err error) {
	logger.Error("%s", err)

	content := map[string]interface{}{
		"Version": a.version,
	}

	var message string
	status := http.StatusInternalServerError

	if err != nil {
		message = err.Error()
		subMessages := ""

		if errors.Is(err, model.ErrInvalid) {
			status = http.StatusBadRequest
		} else if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, model.ErrInternalError) {
			status = http.StatusInternalServerError
			message = "Oops! Something went wrong."
		}

		content["Message"] = model.NewErrorMessage(message)
		if len(subMessages) > 0 {
			content["Errors"] = strings.Split(subMessages, ", ")
		}
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) svg() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.tpl == nil {
			httperror.NotFound(w)
			return
		}

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
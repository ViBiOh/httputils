package renderer

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/sha"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
)

var svgCacheDuration string = fmt.Sprintf("public, max-age=%.0f", time.Duration(time.Minute*10).Seconds())

// Redirect redirect user to a defined path with a message
func (a App) Redirect(w http.ResponseWriter, r *http.Request, pathname string, message Message) {
	joinChar := "?"
	if strings.Contains(pathname, "?") {
		joinChar = "&"
	}

	var anchor string
	parts := strings.SplitN(pathname, "#", 2)
	if len(parts) == 2 && len(parts[1]) > 0 {
		anchor = "#" + parts[1]
	}

	http.Redirect(w, r, fmt.Sprintf("%s%s%s%s", a.url(parts[0]), joinChar, message, anchor), http.StatusFound)
}

func (a App) Error(w http.ResponseWriter, r *http.Request, content map[string]interface{}, err error) {
	logger.Error("%s", err)

	content = a.feedContent(content)

	status, message := httperror.ErrorStatus(err)
	if len(message) > 0 {
		content["Message"] = NewErrorMessage(message)
	}

	nonce := owasp.Nonce()
	owasp.WriteNonce(w, nonce)
	content["nonce"] = nonce

	if err = templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a App) render(w http.ResponseWriter, r *http.Request, templateFunc TemplateFunc) {
	defer func() {
		if exception := recover(); exception != nil {
			output := make([]byte, 1024)
			runtime.Stack(output, false)
			logger.Error("recovered from panic: %s\n%s", exception, output)

			a.Error(w, r, nil, fmt.Errorf("recovered from panic: %s", exception))
		}
	}()

	templateName, status, content, err := templateFunc(w, r)
	if err != nil {
		a.Error(w, r, content, err)
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

	if matchEtag(w, r, content) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	responder := templates.ResponseHTMLTemplate
	if !a.minify {
		responder = templates.ResponseHTMLTemplateRaw
	}

	if err = responder(a.tpl.Lookup(templateName), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func matchEtag(w http.ResponseWriter, r *http.Request, content map[string]interface{}) bool {
	etag := sha.New(content)

	noneMatch := r.Header.Get("If-None-Match")
	if len(noneMatch) == 0 {
		appendNonceAndEtag(w, content, etag)
		return false
	}

	parts := strings.SplitN(noneMatch, "-", 2)
	if len(parts) != 2 {
		appendNonceAndEtag(w, content, etag)
		return false
	}

	if strings.TrimPrefix(parts[0], `W/"`) == etag {
		owasp.WriteNonce(w, strings.TrimSuffix(parts[1], `"`))
		return true
	}

	appendNonceAndEtag(w, content, etag)
	return false
}

func appendNonceAndEtag(w http.ResponseWriter, content map[string]interface{}, etag string) {
	nonce := owasp.Nonce()
	owasp.WriteNonce(w, nonce)
	content["nonce"] = nonce
	w.Header().Add("Etag", fmt.Sprintf(`W/"%s-%s"`, etag, nonce))
}

func (a App) svg() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup("svg-" + strings.Trim(r.URL.Path, "/"))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Add("Cache-Control", svgCacheDuration)
		w.Header().Add("Content-Type", "image/svg+xml")

		if err := templates.WriteTemplate(tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

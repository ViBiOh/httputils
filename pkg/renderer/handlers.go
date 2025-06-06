package renderer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type errOption struct {
	noLog bool
}

type ErrorOption func(errOption) errOption

func WithNoLog() ErrorOption {
	return func(in errOption) errOption {
		in.noLog = true

		return in
	}
}

func (s *Service) Redirect(w http.ResponseWriter, r *http.Request, pathname string, message Message) {
	joinChar := "?"
	if strings.Contains(pathname, "?") {
		joinChar = "&"
	}

	var anchor string

	anchorIndex := strings.Index(pathname, "#")
	if anchorIndex != -1 && len(pathname[anchorIndex+1:]) > 0 {
		anchor = "#" + pathname[anchorIndex+1:]
		pathname = pathname[:anchorIndex]
	}

	http.Redirect(w, r, fmt.Sprintf("%s%s%s%s", s.url(pathname), joinChar, message, anchor), http.StatusFound)
}

func (s *Service) Error(w http.ResponseWriter, r *http.Request, content map[string]any, err error, opts ...ErrorOption) {
	ctx := r.Context()

	if r.Header.Get("Accept") == "text/plain" && httperror.HandleError(ctx, w, err) {
		return
	}

	content = s.feedContent(content)

	var config errOption
	for _, opt := range opts {
		config = opt(config)
	}

	status, statusMessage := httperror.ErrorStatus(err)
	if len(content) > 0 {
		message := NewErrorMessage("%s", statusMessage)
		content[message.Key] = message
	}

	if !config.noLog {
		httperror.Log(ctx, err, status, statusMessage)
	}

	nonce := owasp.Nonce()
	owasp.WriteNonce(w, nonce)
	content["nonce"] = nonce

	if err = templates.ResponseHTMLTemplate(ctx, s.tracer, s.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(ctx, w, err)
	}
}

func (s *Service) render(w http.ResponseWriter, r *http.Request, templateFunc TemplateFunc) {
	defer recoverer.Handler(func(err error) {
		s.Error(w, r, nil, err)
	})

	page, err := templateFunc(w, r)
	if err != nil {
		s.Error(w, r, page.Content, err)

		return
	}

	if len(page.Template) == 0 {
		return
	}

	tpl := s.tpl.Lookup(page.Template)
	if tpl == nil {
		s.Error(w, r, page.Content, model.WrapNotFound(fmt.Errorf("unknown template `%s`", page.Template)))
		return
	}

	page.Content = s.feedContent(page.Content)

	message := ParseMessage(r)
	if len(message.Content) > 0 {
		page.Content[message.Key] = message
	}

	if s.matchEtag(w, r, page) {
		w.WriteHeader(http.StatusNotModified)

		return
	}

	responder := templates.ResponseHTMLTemplate
	if !s.minify {
		responder = templates.ResponseHTMLTemplateRaw
	}

	ctx := r.Context()

	if s.generatedMeter != nil {
		s.generatedMeter.Add(ctx, 1, metric.WithAttributes(attribute.String("template", page.Template)))
	}

	if err = responder(ctx, s.tracer, tpl, w, page.Content, page.Status); err != nil {
		httperror.InternalServerError(ctx, w, err)
	}
}

func (s *Service) matchEtag(w http.ResponseWriter, r *http.Request, page Page) bool {
	etag := page.etag()

	noneMatch := r.Header.Get("If-None-Match")
	if len(noneMatch) == 0 {
		appendNonceAndEtag(w, page.Content, etag)

		return false
	}

	dashIndex := strings.Index(noneMatch, "-")
	if dashIndex == -1 {
		appendNonceAndEtag(w, page.Content, etag)

		return false
	}

	if strings.TrimPrefix(noneMatch[:dashIndex], `W/"`) == etag {
		owasp.WriteNonce(w, strings.TrimSuffix(noneMatch[dashIndex+1:], `"`))

		return true
	}

	appendNonceAndEtag(w, page.Content, etag)

	return false
}

func appendNonceAndEtag(w http.ResponseWriter, content map[string]any, etag string) {
	nonce := owasp.Nonce()
	owasp.WriteNonce(w, nonce)
	content["nonce"] = nonce
	w.Header().Add("Etag", fmt.Sprintf(`W/"%s-%s"`, etag, nonce))
}

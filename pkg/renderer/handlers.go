package renderer

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func (s Service) Redirect(w http.ResponseWriter, r *http.Request, pathname string, message Message) {
	joinChar := "?"
	if strings.Contains(pathname, "?") {
		joinChar = "&"
	}

	var anchor string

	parts := strings.SplitN(pathname, "#", 2)
	if len(parts) == 2 && len(parts[1]) > 0 {
		anchor = "#" + parts[1]
	}

	http.Redirect(w, r, fmt.Sprintf("%s%s%s%s", s.url(parts[0]), joinChar, message, anchor), http.StatusFound)
}

func (s Service) Error(w http.ResponseWriter, r *http.Request, content map[string]any, err error) {
	slog.ErrorContext(r.Context(), err.Error())

	content = s.feedContent(content)

	status, message := httperror.ErrorStatus(err)
	if len(message) > 0 {
		content["Message"] = NewErrorMessage(message)
	}

	nonce := owasp.Nonce()
	owasp.WriteNonce(w, nonce)
	content["nonce"] = nonce

	if err = templates.ResponseHTMLTemplate(r.Context(), s.tracer, s.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(r.Context(), w, err)
	}
}

func (s Service) render(w http.ResponseWriter, r *http.Request, templateFunc TemplateFunc) {
	defer func() {
		if exception := recover(); exception != nil {
			output := make([]byte, 1024)
			runtime.Stack(output, false)
			slog.ErrorContext(r.Context(), "recovered from panic", "err", exception, "stacktrace", string(output))

			s.Error(w, r, nil, fmt.Errorf("recovered from panic: %s", exception))
		}
	}()

	page, err := templateFunc(w, r)
	if err != nil {
		s.Error(w, r, page.Content, err)

		return
	}

	if len(page.Template) == 0 {
		return
	}

	page.Content = s.feedContent(page.Content)

	message := ParseMessage(r)
	if len(message.Content) > 0 {
		page.Content["Message"] = message
	}

	if s.matchEtag(w, r, page) {
		w.WriteHeader(http.StatusNotModified)

		return
	}

	responder := templates.ResponseHTMLTemplate
	if !s.minify {
		responder = templates.ResponseHTMLTemplateRaw
	}

	if s.generatedMeter != nil {
		s.generatedMeter.Add(r.Context(), 1, metric.WithAttributes(attribute.String("template", page.Template)))
	}

	if err = responder(r.Context(), s.tracer, s.tpl.Lookup(page.Template), w, page.Content, page.Status); err != nil {
		httperror.InternalServerError(r.Context(), w, err)
	}
}

func (s Service) matchEtag(w http.ResponseWriter, r *http.Request, page Page) bool {
	_, end := telemetry.StartSpan(r.Context(), s.tracer, "match_etag", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(nil)

	etag := page.etag()

	noneMatch := r.Header.Get("If-None-Match")
	if len(noneMatch) == 0 {
		appendNonceAndEtag(w, page.Content, etag)

		return false
	}

	parts := strings.SplitN(noneMatch, "-", 2)
	if len(parts) != 2 {
		appendNonceAndEtag(w, page.Content, etag)

		return false
	}

	if strings.TrimPrefix(parts[0], `W/"`) == etag {
		owasp.WriteNonce(w, strings.TrimSuffix(parts[1], `"`))

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

func (s Service) svg() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := s.tpl.Lookup("svg-" + strings.Trim(r.URL.Path, "/"))
		if tpl == nil {
			httperror.NotFound(r.Context(), w)

			return
		}

		w.Header().Add("Cache-Control", staticCacheDuration)
		w.Header().Add("Content-Type", "image/svg+xml")

		if err := templates.WriteTemplate(r.Context(), s.tracer, tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(r.Context(), w, err)
		}
	})
}

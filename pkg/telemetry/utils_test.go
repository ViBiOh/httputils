package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	noopmeter "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	tr "go.opentelemetry.io/otel/trace"
)

func newInstrumentedClient(t *testing.T, transport http.RoundTripper) (*http.Client, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})

	return AddOpenTelemetryToClient(&http.Client{Transport: transport}, noopmeter.NewMeterProvider(), tracerProvider), recorder
}

func clientSpan(t *testing.T, recorder *tracetest.SpanRecorder) sdktrace.ReadOnlySpan {
	t.Helper()

	var found []sdktrace.ReadOnlySpan

	for _, span := range recorder.Ended() {
		if span.SpanKind() == tr.SpanKindClient {
			found = append(found, span)
		}
	}

	require.Len(t, found, 1)

	return found[0]
}

func attributeValue(span sdktrace.ReadOnlySpan, key string) string {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}

	return ""
}

func get(ctx context.Context, client *http.Client, target string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func TestRedactedURLInClientSpan(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	secretURL := server.URL + "/bot123456:SECRET-TOKEN/sendMessage?chat_id=SECRET-CHAT"

	t.Run("redacted", func(t *testing.T) {
		t.Parallel()

		client, recorder := newInstrumentedClient(t, nil)

		resp, err := get(WithRedactedURL(context.Background()), client, secretURL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		span := clientSpan(t, recorder)

		assert.Equal(t, server.URL+"/<redacted>", attributeValue(span, "url.full"))
		assert.NotContains(t, attributeValue(span, "url.full"), "SECRET")

		for _, attr := range span.Attributes() {
			assert.NotContains(t, attr.Value.String(), "SECRET", string(attr.Key))
		}

		assert.Equal(t, int64(http.StatusNoContent), findInt(span, "http.response.status_code"))
	})

	t.Run("not redacted by default", func(t *testing.T) {
		t.Parallel()

		client, recorder := newInstrumentedClient(t, nil)

		resp, err := get(context.Background(), client, secretURL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		assert.Equal(t, secretURL, attributeValue(clientSpan(t, recorder), "url.full"))
	})

	t.Run("only the marked requests", func(t *testing.T) {
		t.Parallel()

		client, recorder := newInstrumentedClient(t, nil)

		resp, err := get(WithRedactedURL(context.Background()), client, secretURL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		resp, err = get(context.Background(), client, server.URL+"/public/path")
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		var urls []string

		for _, span := range recorder.Ended() {
			urls = append(urls, attributeValue(span, "url.full"))
		}

		assert.ElementsMatch(t, []string{server.URL + "/<redacted>", server.URL + "/public/path"}, urls)
	})

	t.Run("failed request", func(t *testing.T) {
		t.Parallel()

		client, recorder := newInstrumentedClient(t, nil)

		resp, err := get(WithRedactedURL(context.Background()), client, "http://127.0.0.1:1/bot123456:SECRET-TOKEN/sendMessage")
		if resp != nil {
			require.NoError(t, resp.Body.Close())
		}

		require.Error(t, err)

		span := clientSpan(t, recorder)

		assert.Equal(t, codes.Error, span.Status().Code)
		assert.NotContains(t, span.Status().Description, "SECRET")
		assert.Equal(t, "http://127.0.0.1:1/<redacted>", attributeValue(span, "url.full"))

		for _, event := range span.Events() {
			for _, attr := range event.Attributes {
				assert.NotContains(t, attr.Value.String(), "SECRET", event.Name)
			}
		}
	})

	t.Run("uses the given transport", func(t *testing.T) {
		t.Parallel()

		var called bool

		client, recorder := newInstrumentedClient(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			called = true

			assert.Equal(t, secretURL, req.URL.String(), "only the span is redacted, not the request")

			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
		}))

		resp, err := get(WithRedactedURL(context.Background()), client, secretURL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		assert.True(t, called)
		assert.Equal(t, server.URL+"/<redacted>", attributeValue(clientSpan(t, recorder), "url.full"))
	})
}

func TestRedactURL(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input string
		want  string
	}{
		"path and query": {
			"https://api.telegram.org/bot123456:SECRET/sendMessage?chat_id=42",
			"https://api.telegram.org/<redacted>",
		},
		"path only": {
			"https://hooks.slack.com/services/T000/B000/SECRET",
			"https://hooks.slack.com/<redacted>",
		},
		"query only": {
			"https://example.com?token=SECRET",
			"https://example.com/<redacted>",
		},
		"credentials and port": {
			"https://user:SECRET@example.com:8443/hook",
			"https://example.com:8443/<redacted>",
		},
		"host only": {
			"https://example.com",
			"https://example.com",
		},
		"root path": {
			"https://example.com/",
			"https://example.com",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			parsed, err := url.Parse(tc.input)
			require.NoError(t, err)

			assert.Equal(t, tc.want, RedactURL(parsed))
		})
	}

	assert.Equal(t, "<redacted>", RedactURL(nil))
	assert.Equal(t, "<redacted>", RedactURL(&url.URL{Path: "/SECRET"}))
}

func TestAddOpenTelemetryToClientWithoutProvider(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	assert.Same(t, client, AddOpenTelemetryToClient(client, noopmeter.NewMeterProvider(), nil))
	assert.Nil(t, client.Transport)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func findInt(span sdktrace.ReadOnlySpan, key string) int64 {
	for _, attr := range span.Attributes() {
		if attr.Key == attribute.Key(key) {
			return attr.Value.AsInt64()
		}
	}

	return 0
}

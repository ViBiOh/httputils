package renderer

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"github.com/ViBiOh/httputils/v4/pkg/hash"
)

const (
	defaultMessageKey = "Message"
	paramKey          = "msgKey"
	paramLevel        = "msgLvl"
	paramContent      = "msgCnt"
)

type TemplateFunc = func(http.ResponseWriter, *http.Request) (Page, error)

type Page struct {
	Content  map[string]any
	Template string
	Status   int
}

func NewPage(template string, status int, content map[string]any) Page {
	return Page{
		Template: template,
		Status:   status,
		Content:  content,
	}
}

func (p Page) etag() string {
	streamer := hash.Stream()

	streamer.WriteString(p.Template)
	streamer.Write(p.Status)

	keys := make([]string, 0, len(p.Content))
	keys = slices.AppendSeq(keys, maps.Keys(p.Content))
	slices.Sort(keys)

	for _, key := range keys {
		streamer.WriteString(key)

		switch typedValue := p.Content[key].(type) {
		case string:
			streamer.WriteString(typedValue)
		case []byte:
			streamer.WriteBytes(typedValue)
		default:
			streamer.Write(typedValue)
		}
	}

	return streamer.Sum()
}

type Message struct {
	Key     string
	Level   string
	Content string
}

func newMessage(key, level, format string, a ...any) Message {
	return Message{
		Key:     key,
		Level:   level,
		Content: fmt.Sprintf(format, a...),
	}
}

func (m Message) String() string {
	if len(m.Content) == 0 {
		return ""
	}

	return fmt.Sprintf("%s=%s&%s=%s&%s=%s", paramKey, url.QueryEscape(m.Key), paramContent, url.QueryEscape(m.Content), paramLevel, url.QueryEscape(m.Level))
}

func ParseMessage(r *http.Request) Message {
	values := r.URL.Query()

	return Message{
		Key:     values.Get(paramKey),
		Level:   values.Get(paramLevel),
		Content: values.Get(paramContent),
	}
}

func NewSuccessMessage(format string, a ...any) Message {
	return newMessage(defaultMessageKey, "success", format, a...)
}

func NewErrorMessage(format string, a ...any) Message {
	return newMessage(defaultMessageKey, "error", format, a...)
}

func NewKeyErrorMessage(key, format string, a ...any) Message {
	return newMessage(key, "error", format, a...)
}

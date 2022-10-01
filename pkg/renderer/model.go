package renderer

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v4/pkg/sha"
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
	return sha.New(p)
}

type Message struct {
	Level   string
	Content string
}

func newMessage(level, format string, a ...any) Message {
	return Message{
		Level:   level,
		Content: fmt.Sprintf(format, a...),
	}
}

func (m Message) String() string {
	if len(m.Content) == 0 {
		return ""
	}

	return fmt.Sprintf("messageContent=%s&messageLevel=%s", url.QueryEscape(m.Content), url.QueryEscape(m.Level))
}

func ParseMessage(r *http.Request) Message {
	values := r.URL.Query()

	return Message{
		Level:   values.Get("messageLevel"),
		Content: values.Get("messageContent"),
	}
}

func NewSuccessMessage(format string, a ...any) Message {
	return newMessage("success", format, a...)
}

func NewErrorMessage(format string, a ...any) Message {
	return newMessage("error", format, a...)
}

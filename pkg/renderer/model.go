package renderer

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v4/pkg/sha"
)

// TemplateFunc handle a request and returns which template to render with which status and datas
type TemplateFunc = func(http.ResponseWriter, *http.Request) (Page, error)

// Page describes a page for the renderer
type Page struct {
	Content  map[string]any
	Template string
	Status   int
}

// NewPage creates a new page
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

// Message for render
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

// ParseMessage parses messages from request
func ParseMessage(r *http.Request) Message {
	values := r.URL.Query()

	return Message{
		Level:   values.Get("messageLevel"),
		Content: values.Get("messageContent"),
	}
}

// NewSuccessMessage create a success message
func NewSuccessMessage(format string, a ...any) Message {
	return newMessage("success", format, a...)
}

// NewErrorMessage create a error message
func NewErrorMessage(format string, a ...any) Message {
	return newMessage("error", format, a...)
}

package renderer

import (
	"fmt"
	"net/http"
	"net/url"
)

// TemplateFunc handle a request and returns which template to render with which status and datas
type TemplateFunc = func(*http.Request) (string, int, map[string]interface{}, error)

// Message for render
type Message struct {
	Level   string
	Content string
}

func newMessage(level, format string, a ...interface{}) Message {
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
func NewSuccessMessage(format string, a ...interface{}) Message {
	return newMessage("success", format, a...)
}

// NewErrorMessage create a error message
func NewErrorMessage(format string, a ...interface{}) Message {
	return newMessage("error", format, a...)
}

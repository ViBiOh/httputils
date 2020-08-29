package logger

import (
	"bytes"
	"strings"
	"time"
)

type event struct {
	timestamp time.Time
	level     level
	message   string
}

func (e event) json(builder *bytes.Buffer) []byte {
	builder.Reset()

	builder.WriteString(`{"timestamp":`)
	builder.WriteString(e.timestamp.Format(time.RFC3339))
	builder.WriteString(`,"level":"`)
	builder.WriteString(levelValues[e.level])
	builder.WriteString(`","message":"`)
	builder.WriteString(EscapeString(e.message))
	builder.WriteString(`"}`)
	builder.WriteString("\n")

	return builder.Bytes()
}

func (e event) text(builder *bytes.Buffer) []byte {
	builder.Reset()

	builder.WriteString(e.timestamp.Format(time.RFC3339))
	builder.WriteString(` `)
	builder.WriteString(levelValues[e.level])
	builder.WriteString(` `)
	builder.WriteString(e.message)
	builder.WriteString("\n")

	return builder.Bytes()
}

// EscapeString escapes value from raw string to be JSON compatible
func EscapeString(content string) string {
	output := content

	if strings.Contains(output, "\\") {
		output = strings.ReplaceAll(output, "\\", "\\\\")
	}

	if strings.Contains(output, "\b") {
		output = strings.ReplaceAll(output, "\b", "\\b")
	}

	if strings.Contains(output, "\f") {
		output = strings.ReplaceAll(output, "\f", "\\f")
	}

	if strings.Contains(output, "\r") {
		output = strings.ReplaceAll(output, "\r", "\\r")
	}

	if strings.Contains(output, "\n") {
		output = strings.ReplaceAll(output, "\n", "\\n")
	}

	if strings.Contains(output, "\t") {
		output = strings.ReplaceAll(output, "\t", "\\t")
	}

	if strings.Contains(output, "\"") {
		output = strings.ReplaceAll(output, "\"", "\\\"")
	}

	return output
}

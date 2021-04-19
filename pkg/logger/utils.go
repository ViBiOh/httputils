package logger

import (
	"bytes"
	"strings"
)

// EscapeString escapes value from raw string to be JSON compatible
func EscapeString(content string) string {
	if !strings.ContainsRune(content, '\\') && !strings.ContainsRune(content, '\b') && !strings.ContainsRune(content, '\f') && !strings.ContainsRune(content, '\r') && !strings.ContainsRune(content, '\n') && !strings.ContainsRune(content, '\t') && !strings.ContainsRune(content, '"') {
		return content
	}

	output := bytes.NewBuffer(nil)

	for _, char := range content {
		switch char {
		case '\\':
			output.WriteString(`\\`)
		case '\b':
			output.WriteString(`\b`)
		case '\f':
			output.WriteString(`\f`)
		case '\r':
			output.WriteString(`\r`)
		case '\n':
			output.WriteString(`\n`)
		case '\t':
			output.WriteString(`\t`)
		case '"':
			output.WriteString(`\"`)
		default:
			output.WriteRune(char)
		}
	}

	return output.String()
}

package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type level int

const (
	levelFatal = iota
	levelError
	levelWarning
	levelInfo
	levelDebug
	levelTrace
)

var (
	levelValues = []string{"FATAL", "ERROR", "WARNING", "INFO", "DEBUG", "TRACE"}
)

func parseLevel(line string) (level, error) {
	for i, l := range levelValues {
		if strings.EqualFold(l, line) {
			return level(i), nil
		}
	}

	return levelInfo, fmt.Errorf("invalid value `%s` for level", line)
}

type event struct {
	timestamp time.Time
	message   string
	level     level
}

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

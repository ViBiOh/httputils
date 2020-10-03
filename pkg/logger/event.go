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
	level     level
	message   string
}

// EscapeString escapes value from raw string to be JSON compatible
func EscapeString(content string) string {
	if !strings.ContainsRune(content, '\\') && !strings.ContainsRune(content, '\b') && !strings.ContainsRune(content, '\f') && !strings.ContainsRune(content, '\r') && !strings.ContainsRune(content, '\n') && !strings.ContainsRune(content, '\t') && !strings.ContainsRune(content, '"') {
		return content
	}

	output := bytes.Buffer{}

	for _, char := range content {
		switch char {
		case '\\':
			output.WriteString(`\\`)
			break
		case '\b':
			output.WriteString(`\b`)
			break
		case '\f':
			output.WriteString(`\f`)
			break
		case '\r':
			output.WriteString(`\r`)
			break
		case '\n':
			output.WriteString(`\n`)
			break
		case '\t':
			output.WriteString(`\t`)
			break
		case '"':
			output.WriteString(`\"`)
			break
		default:
			output.WriteRune(char)
		}
	}

	return output.String()
}

package logger

import (
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
	levelValues = []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}
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

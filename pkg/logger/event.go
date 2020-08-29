package logger

import (
	"strings"
	"time"
)

type event struct {
	timestamp time.Time
	level     level
	message   string
}

func (e event) json(logger *Logger) []byte {
	logger.builder.Reset()

	logger.builder.WriteString(`{"`)
	logger.builder.WriteString(logger.timeKey)
	logger.builder.WriteString(`":"`)
	logger.builder.WriteString(e.timestamp.Format(time.RFC3339))
	logger.builder.WriteString(`","`)
	logger.builder.WriteString(logger.levelKey)
	logger.builder.WriteString(`":"`)
	logger.builder.WriteString(levelValues[e.level])
	logger.builder.WriteString(`","`)
	logger.builder.WriteString(logger.messageKey)
	logger.builder.WriteString(`":"`)
	logger.builder.WriteString(EscapeString(e.message))
	logger.builder.WriteString(`"}`)
	logger.builder.WriteString("\n")

	return logger.builder.Bytes()
}

func (e event) text(logger *Logger) []byte {
	logger.builder.Reset()

	logger.builder.WriteString(e.timestamp.Format(time.RFC3339))
	logger.builder.WriteString(` `)
	logger.builder.WriteString(levelValues[e.level])
	logger.builder.WriteString(` `)
	logger.builder.WriteString(e.message)
	logger.builder.WriteString("\n")

	return logger.builder.Bytes()
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

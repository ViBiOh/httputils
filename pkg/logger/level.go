package logger

import (
	"fmt"
	"strings"
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

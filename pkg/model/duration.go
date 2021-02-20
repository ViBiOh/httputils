package model

import (
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// SafeParseDuration parses given value into duration and fallback to a default value
func SafeParseDuration(name string, value string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		logger.Warn("invalid %s value `%s`: %s", name, value, err)
		return defaultDuration
	}

	return duration
}

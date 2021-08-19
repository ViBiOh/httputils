package model

import (
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// SafeParseDuration parses given value into duration and fallback to a default value
func SafeParseDuration(name string, value string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.WithField("name", name).Warn("invalid value `%s`: %s", value, err)
		return defaultDuration
	}

	return duration
}

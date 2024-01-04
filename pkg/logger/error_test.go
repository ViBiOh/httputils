package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/stretchr/testify/assert"
)

type entry struct {
	Level   string            `json:"level"`
	Message string            `json:"msg"`
	Error   map[string]string `json:"error"`
}

func TestErrorField(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expected := entry{
			Level:   slog.LevelError.String(),
			Message: "simple test",
			Error: map[string]string{
				"kind":    "*errors.errorString",
				"message": "boom!",
			},
		}

		logOutput := bytes.NewBuffer(nil)

		logger := configureLogger(logOutput, slog.LevelInfo, true, "time", "level", "msg")

		logger.ErrorContext(context.Background(), "simple test", "error", errors.New("boom!"))

		var got entry
		err := json.Unmarshal(logOutput.Bytes(), &got)
		assert.NoError(t, err)

		assert.Empty(t, got.Error["stack"])
		assert.Equal(t, expected, got)
	})

	t.Run("stacktrace", func(t *testing.T) {
		t.Parallel()

		expected := entry{
			Level:   slog.LevelError.String(),
			Message: "simple test",
			Error: map[string]string{
				"kind":    "recoverer.errWithStackTrace",
				"message": "boom!",
			},
		}

		logOutput := bytes.NewBuffer(nil)

		logger := configureLogger(logOutput, slog.LevelInfo, true, "time", "level", "msg")

		logger.ErrorContext(context.Background(), "simple test", "error", recoverer.WithStack(errors.New("boom!")))

		var got entry
		err := json.Unmarshal(logOutput.Bytes(), &got)
		assert.NoError(t, err)

		assert.NotEmpty(t, got.Error["stack"])
		delete(got.Error, "stack")

		assert.Equal(t, expected, got)
	})

	t.Run("wrapped stacktrace", func(t *testing.T) {
		t.Parallel()

		expected := entry{
			Level:   slog.LevelError.String(),
			Message: "simple test",
			Error: map[string]string{
				"kind":    "*errors.joinError",
				"message": "root\nnested",
			},
		}

		logOutput := bytes.NewBuffer(nil)

		logger := configureLogger(logOutput, slog.LevelInfo, true, "time", "level", "msg")

		logger.ErrorContext(context.Background(), "simple test", "error", errors.Join(errors.New("root"), recoverer.WithStack(errors.New("nested"))))

		var got entry
		err := json.Unmarshal(logOutput.Bytes(), &got)
		assert.NoError(t, err)

		assert.NotEmpty(t, got.Error["stack"])
		delete(got.Error, "stack")

		assert.Equal(t, expected, got)
	})
}

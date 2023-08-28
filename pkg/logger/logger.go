package logger

import (
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/ViBiOh/flags"
)

type Config struct {
	Level      string
	TimeKey    string
	LevelKey   string
	MessageKey string
	JSON       bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Level", "Logger level").Prefix(prefix).DocPrefix("logger").StringVar(fs, &config.Level, "INFO", overrides)
	flags.New("Json", "Log format as JSON").Prefix(prefix).DocPrefix("logger").BoolVar(fs, &config.JSON, false, overrides)
	flags.New("TimeKey", "Key for timestamp in JSON").Prefix(prefix).DocPrefix("logger").StringVar(fs, &config.TimeKey, "time", overrides)
	flags.New("LevelKey", "Key for level in JSON").Prefix(prefix).DocPrefix("logger").StringVar(fs, &config.LevelKey, "level", overrides)
	flags.New("MessageKey", "Key for message in JSON").Prefix(prefix).DocPrefix("logger").StringVar(fs, &config.MessageKey, "msg", overrides)

	return &config
}

func init() {
	configureLogger(os.Stdout, slog.LevelInfo, false, "time", "level", "msg")
}

func Init(config *Config) {
	var level slog.Level

	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		slog.Error(err.Error(), "level", config.Level)

		return
	}

	configureLogger(os.Stdout, level, config.JSON, config.TimeKey, config.LevelKey, config.MessageKey)
}

func configureLogger(writer io.Writer, level slog.Level, json bool, timeKey, levelKey, messageKey string) {
	options := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case "msg":
				a.Key = messageKey
			case "level":
				a.Key = levelKey
			case "time":
				a.Key = timeKey
			}

			return a
		},
	}

	var handler slog.Handler
	if json {
		handler = slog.NewJSONHandler(writer, options)
	} else {
		handler = slog.NewTextHandler(writer, options)
	}

	slog.SetDefault(slog.New(handler))
}

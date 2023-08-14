package logger

import (
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/ViBiOh/flags"
)

type Config struct {
	level      *string
	json       *bool
	timeKey    *string
	levelKey   *string
	messageKey *string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		level:      flags.New("Level", "Logger level").Prefix(prefix).DocPrefix("logger").String(fs, "INFO", overrides),
		json:       flags.New("Json", "Log format as JSON").Prefix(prefix).DocPrefix("logger").Bool(fs, false, overrides),
		timeKey:    flags.New("TimeKey", "Key for timestamp in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "time", overrides),
		levelKey:   flags.New("LevelKey", "Key for level in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "level", overrides),
		messageKey: flags.New("MessageKey", "Key for message in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "msg", overrides),
	}
}

func init() {
	configureLogger(os.Stdout, slog.LevelInfo, false, "time", "level", "msg")
}

func Init(config Config) {
	var level slog.Level

	if err := level.UnmarshalText([]byte(*config.level)); err != nil {
		slog.Error(err.Error(), "level", *config.level)

		return
	}

	configureLogger(os.Stdout, level, *config.json, *config.timeKey, *config.levelKey, *config.messageKey)
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

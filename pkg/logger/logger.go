package logger

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/ViBiOh/flags"
	"golang.org/x/term"
)

type GetNow func() time.Time

var (
	logger   *Logger
	exitFunc = os.Exit

	colorReset  = []byte("\033[0m")
	colorRed    = []byte("\033[31m")
	colorYellow = []byte("\033[33m")
)

type Config struct {
	level      *string
	json       *bool
	timeKey    *string
	levelKey   *string
	messageKey *string
}

type Logger struct {
	clock GetNow

	timeKey    string
	levelKey   string
	messageKey string

	events chan event
	done   chan struct{}

	outWriter io.Writer
	errWriter io.Writer

	outputBuffer *bytes.Buffer
	dateBuffer   []byte

	level      slog.Level
	jsonFormat bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		level:      flags.New("Level", "Logger level").Prefix(prefix).DocPrefix("logger").String(fs, "INFO", overrides),
		json:       flags.New("Json", "Log format as JSON").Prefix(prefix).DocPrefix("logger").Bool(fs, false, overrides),
		timeKey:    flags.New("TimeKey", "Key for timestamp in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "time", overrides),
		levelKey:   flags.New("LevelKey", "Key for level in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "level", overrides),
		messageKey: flags.New("MessageKey", "Key for message in JSON").Prefix(prefix).DocPrefix("logger").String(fs, "message", overrides),
	}
}

func init() {
	logger = newLogger(os.Stdout, os.Stderr, slog.LevelInfo, false, "time", "level", "message")
	go logger.Start()
}

func New(config Config) *Logger {
	var level slog.Level
	err := level.UnmarshalText([]byte(*config.level))

	logger := newLogger(os.Stdout, os.Stderr, level, *config.json, *config.timeKey, *config.levelKey, *config.messageKey)

	go logger.Start()

	if err != nil {
		logger.Error(err.Error())
	}

	return logger
}

func newLogger(outWriter, errWriter io.Writer, lev slog.Level, json bool, timeKey, levelKey, messageKey string) *Logger {
	return &Logger{
		clock: time.Now,

		done:   make(chan struct{}),
		events: make(chan event, runtime.NumCPU()),

		outputBuffer: bytes.NewBuffer(nil),
		dateBuffer:   make([]byte, 25),

		level:     lev,
		outWriter: outWriter,
		errWriter: errWriter,

		jsonFormat: json,
		timeKey:    timeKey,
		levelKey:   levelKey,
		messageKey: messageKey,
	}
}

func (l *Logger) Start() {
	var payload []byte
	var err error

	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))

	for e := range l.events {
		if l.jsonFormat {
			payload = l.json(e)
		} else {
			payload = l.text(e)
		}

		if isTerminal {
			if color := getColor(e.level); color != nil {
				payload = append(color, payload...)
				payload = append(payload, colorReset...)
			}
		}

		if e.level <= slog.LevelInfo {
			_, err = l.outWriter.Write(payload)
		} else {
			_, err = l.errWriter.Write(payload)
		}

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write log: %s\n", err)
		}
	}

	close(l.done)
}

func getColor(level slog.Level) []byte {
	switch level {
	case slog.LevelWarn:
		return colorYellow
	case slog.LevelError:
		return colorRed
	default:
		return nil
	}
}

func (l *Logger) Enabled(ctx context.Context, lev slog.Level) bool {
	return l.level <= lev
}

func (l *Logger) Close() {
	close(l.events)
	<-l.done
}

func (l *Logger) Debug(format string, a ...any) {
	if l.isIgnored(slog.LevelDebug) {
		return
	}

	l.output(slog.LevelDebug, nil, format, a...)
}

func (l *Logger) Info(format string, a ...any) {
	if l.isIgnored(slog.LevelInfo) {
		return
	}

	l.output(slog.LevelInfo, nil, format, a...)
}

func (l *Logger) Warn(format string, a ...any) {
	if l.isIgnored(slog.LevelWarn) {
		return
	}

	l.output(slog.LevelWarn, nil, format, a...)
}

func (l *Logger) Error(format string, a ...any) {
	if l.isIgnored(slog.LevelError) {
		return
	}

	l.output(slog.LevelError, nil, format, a...)
}

func (l *Logger) Fatal(err error) {
	if err == nil {
		return
	}

	l.output(slog.LevelError, nil, "%s", err)
	l.Close()

	exitFunc(1)
}

func (l *Logger) WithField(name string, value any) Provider {
	return FieldsContext{
		outputFn: l.output,
		closeFn:  l.Close,
		fields: []field{{
			name:  name,
			value: value,
		}},
	}
}

func (l *Logger) isIgnored(lev slog.Level) bool {
	return l.level > lev
}

func (l *Logger) output(lev slog.Level, fields []field, format string, a ...any) {
	if l.isIgnored(lev) {
		return
	}

	if len(a) > 0 {
		format = fmt.Sprintf(format, a...)
	}

	l.events <- event{timestamp: l.clock(), level: lev, message: format, fields: fields}
}

func (l *Logger) json(e event) []byte {
	l.outputBuffer.Reset()

	l.outputBuffer.WriteString(`{"`)
	l.outputBuffer.WriteString(l.timeKey)
	l.outputBuffer.WriteString(`":"`)
	l.outputBuffer.Write(e.timestamp.AppendFormat(l.dateBuffer[:0], time.RFC3339))
	l.outputBuffer.WriteString(`","`)
	l.outputBuffer.WriteString(l.levelKey)
	l.outputBuffer.WriteString(`":"`)
	l.outputBuffer.WriteString(e.level.String())
	l.outputBuffer.WriteString(`","`)
	l.outputBuffer.WriteString(l.messageKey)
	l.outputBuffer.WriteString(`":"`)
	WriteEscapedJSON(e.message, l.outputBuffer)
	l.outputBuffer.WriteString(`"`)

	for _, field := range e.fields {
		l.outputBuffer.WriteString(`,"`)
		WriteEscapedJSON(field.name, l.outputBuffer)
		l.outputBuffer.WriteString(`":`)

		switch content := field.value.(type) {
		case string:
			l.outputBuffer.WriteString(`"`)
			WriteEscapedJSON(content, l.outputBuffer)
			l.outputBuffer.WriteString(`"`)
		default:
			WriteEscapedJSON(fmt.Sprintf("%v", field.value), l.outputBuffer)
		}
	}

	l.outputBuffer.WriteString(`}`)
	l.outputBuffer.WriteString("\n")

	return l.outputBuffer.Bytes()
}

func (l *Logger) text(e event) []byte {
	l.outputBuffer.Reset()

	l.outputBuffer.Write(e.timestamp.AppendFormat(l.dateBuffer[:0], time.RFC3339))
	l.outputBuffer.WriteString(` `)
	l.outputBuffer.WriteString(e.level.String())
	l.outputBuffer.WriteString(` `)
	l.outputBuffer.WriteString(e.message)

	for _, field := range e.fields {
		l.outputBuffer.WriteString(" ")
		l.outputBuffer.WriteString(field.name)
		l.outputBuffer.WriteString("=")
		_, _ = fmt.Fprintf(l.outputBuffer, "%#v", field.value)
	}
	l.outputBuffer.WriteString("\n")

	return l.outputBuffer.Bytes()
}

func (l *Logger) Errorf(format string, a ...any) {
	l.Error(format, a...)
}

func (l *Logger) Warningf(format string, a ...any) {
	l.Warn(format, a...)
}

func (l *Logger) Infof(format string, a ...any) {
	l.Info(format, a...)
}

func (l *Logger) Debugf(format string, a ...any) {
	l.Debug(format, a...)
}

package logger

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"golang.org/x/term"
)

var (
	logger   Logger
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
	clock clock.Clock

	timeKey    string
	levelKey   string
	messageKey string

	events chan event
	done   chan struct{}

	outWriter io.Writer
	errWriter io.Writer

	outputBuffer *bytes.Buffer
	dateBuffer   []byte

	level      level
	jsonFormat bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		level:      flags.String(fs, prefix, "logger", "Level", "Logger level", "INFO", overrides),
		json:       flags.Bool(fs, prefix, "logger", "Json", "Log format as JSON", false, overrides),
		timeKey:    flags.String(fs, prefix, "logger", "TimeKey", "Key for timestamp in JSON", "time", overrides),
		levelKey:   flags.String(fs, prefix, "logger", "LevelKey", "Key for level in JSON", "level", overrides),
		messageKey: flags.String(fs, prefix, "logger", "MessageKey", "Key for message in JSON", "message", overrides),
	}
}

func init() {
	logger = newLogger(os.Stdout, os.Stderr, levelInfo, false, "time", "level", "message")
	go logger.Start()
}

func New(config Config) Logger {
	level, err := parseLevel(*config.level)

	logger := newLogger(os.Stdout, os.Stderr, level, *config.json, *config.timeKey, *config.levelKey, *config.messageKey)

	go logger.Start()

	if err != nil {
		logger.Error(err.Error())
	}

	return logger
}

func newLogger(outWriter, errWriter io.Writer, lev level, json bool, timeKey, levelKey, messageKey string) Logger {
	return Logger{
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

func (l Logger) Start() {
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

		if e.level <= levelInfo {
			_, err = l.outWriter.Write(payload)
		} else {
			_, err = l.errWriter.Write(payload)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "write log: %s\n", err)
		}
	}

	close(l.done)
}

func getColor(level level) []byte {
	switch level {
	case levelWarning:
		return colorYellow
	case levelError, levelFatal:
		return colorRed
	default:
		return nil
	}
}

func (l Logger) Close() {
	close(l.events)
	<-l.done
}

func (l Logger) Trace(format string, a ...any) {
	l.output(levelTrace, nil, format, a...)
}

func (l Logger) Debug(format string, a ...any) {
	l.output(levelDebug, nil, format, a...)
}

func (l Logger) Info(format string, a ...any) {
	l.output(levelInfo, nil, format, a...)
}

func (l Logger) Warn(format string, a ...any) {
	l.output(levelWarning, nil, format, a...)
}

func (l Logger) Error(format string, a ...any) {
	l.output(levelError, nil, format, a...)
}

func (l Logger) Fatal(err error) {
	if err == nil {
		return
	}

	l.output(levelFatal, nil, "%s", err)
	l.Close()

	exitFunc(1)
}

func (l Logger) WithField(name string, value any) Provider {
	return FieldsContext{
		outputFn: l.output,
		closeFn:  l.Close,
		fields: []field{{
			name:  name,
			value: value,
		}},
	}
}

func (l Logger) output(lev level, fields []field, format string, a ...any) {
	if l.level < lev {
		return
	}

	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}

	l.events <- event{timestamp: l.clock.Now(), level: lev, message: message, fields: fields}
}

func (l Logger) json(e event) []byte {
	l.outputBuffer.Reset()

	l.outputBuffer.WriteString(`{"`)
	l.outputBuffer.WriteString(l.timeKey)
	l.outputBuffer.WriteString(`":"`)
	l.outputBuffer.Write(e.timestamp.AppendFormat(l.dateBuffer[:0], time.RFC3339))
	l.outputBuffer.WriteString(`","`)
	l.outputBuffer.WriteString(l.levelKey)
	l.outputBuffer.WriteString(`":"`)
	l.outputBuffer.WriteString(levelValues[e.level])
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

func (l Logger) text(e event) []byte {
	l.outputBuffer.Reset()

	l.outputBuffer.Write(e.timestamp.AppendFormat(l.dateBuffer[:0], time.RFC3339))
	l.outputBuffer.WriteString(` `)
	l.outputBuffer.WriteString(levelValues[e.level])
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

func (l Logger) Errorf(format string, a ...any) {
	l.Error(format, a...)
}

func (l Logger) Warningf(format string, a ...any) {
	l.Warn(format, a...)
}

func (l Logger) Infof(format string, a ...any) {
	l.Info(format, a...)
}

func (l Logger) Debugf(format string, a ...any) {
	l.Debug(format, a...)
}

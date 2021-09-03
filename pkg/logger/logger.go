package logger

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
)

var (
	logger   Logger
	exitFunc = os.Exit
)

// Config of package
type Config struct {
	level      *string
	json       *bool
	timeKey    *string
	levelKey   *string
	messageKey *string
}

// Logger defines a logger instance
type Logger struct {
	clock *clock.Clock

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

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		level:      flags.New(prefix, "logger", "Level").Default("INFO", overrides).Label("Logger level").ToString(fs),
		json:       flags.New(prefix, "logger", "Json").Default(false, overrides).Label("Log format as JSON").ToBool(fs),
		timeKey:    flags.New(prefix, "logger", "TimeKey").Default("time", overrides).Label("Key for timestamp in JSON").ToString(fs),
		levelKey:   flags.New(prefix, "logger", "LevelKey").Default("level", overrides).Label("Key for level in JSON").ToString(fs),
		messageKey: flags.New(prefix, "logger", "MessageKey").Default("message", overrides).Label("Key for message in JSON").ToString(fs),
	}
}

func init() {
	logger = newLogger(os.Stdout, os.Stderr, levelInfo, false, "time", "level", "message")
	go logger.Start()
}

// New creates a Logger
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

// Start starts logger's writer
func (l Logger) Start() {
	var payload []byte
	var err error

	for e := range l.events {
		if l.jsonFormat {
			payload = l.json(e)
		} else {
			payload = l.text(e)
		}

		if e.level <= levelInfo {
			_, err = l.outWriter.Write(payload)
		} else {
			_, err = l.errWriter.Write(payload)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write log: %s\n", err)
		}
	}

	close(l.done)
}

// Close ends logger gracefully
func (l Logger) Close() {
	close(l.events)
	<-l.done
}

// Trace logs tracing message
func (l Logger) Trace(format string, a ...interface{}) {
	l.output(levelTrace, nil, format, a...)
}

// Debug logs debug message
func (l Logger) Debug(format string, a ...interface{}) {
	l.output(levelDebug, nil, format, a...)
}

// Info logs info message
func (l Logger) Info(format string, a ...interface{}) {
	l.output(levelInfo, nil, format, a...)
}

// Warn logs warning message
func (l Logger) Warn(format string, a ...interface{}) {
	l.output(levelWarning, nil, format, a...)
}

// Error logs error message
func (l Logger) Error(format string, a ...interface{}) {
	l.output(levelError, nil, format, a...)
}

// Fatal logs error message and exit with status code 1
func (l Logger) Fatal(err error) {
	if err == nil {
		return
	}

	l.output(levelFatal, nil, "%s", err)
	l.Close()

	exitFunc(1)
}

// WithField add given name and value to a context
func (l Logger) WithField(name string, value interface{}) FieldsContext {
	return FieldsContext{
		outputFn: l.output,
		closeFn:  l.Close,
		fields: []field{{
			name:  name,
			value: value,
		}},
	}
}

func (l Logger) output(lev level, fields []field, format string, a ...interface{}) {
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
		fmt.Fprintf(l.outputBuffer, "%#v", field.value)
	}
	l.outputBuffer.WriteString("\n")

	return l.outputBuffer.Bytes()
}

// Providing function wrapper for interface compatibility

// Errorf logs error message
func (l Logger) Errorf(format string, a ...interface{}) {
	l.Error(format, a...)
}

// Warningf logs warning message
func (l Logger) Warningf(format string, a ...interface{}) {
	l.Warn(format, a...)
}

// Infof logs info message
func (l Logger) Infof(format string, a ...interface{}) {
	l.Info(format, a...)
}

// Debugf logs debug message
func (l Logger) Debugf(format string, a ...interface{}) {
	l.Debug(format, a...)
}

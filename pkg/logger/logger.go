package logger

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
)

var (
	logger   Logger
	exitFunc = os.Exit
	nowFunc  = time.Now
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
		level:      flags.New(prefix, "logger").Name("Level").Default(flags.Default("Level", "INFO", overrides)).Label("Logger level").ToString(fs),
		json:       flags.New(prefix, "logger").Name("Json").Default(flags.Default("Json", false, overrides)).Label("Log format as JSON").ToBool(fs),
		timeKey:    flags.New(prefix, "logger").Name("TimeKey").Default(flags.Default("TimeKey", "time", overrides)).Label("Key for timestamp in JSON").ToString(fs),
		levelKey:   flags.New(prefix, "logger").Name("LevelKey").Default(flags.Default("LevelKey", "level", overrides)).Label("Key for level in JSON").ToString(fs),
		messageKey: flags.New(prefix, "logger").Name("MessageKey").Default(flags.Default("MessageKey", "message", overrides)).Label("Key for message in JSON").ToString(fs),
	}
}

func init() {
	logger = Logger{
		done:   make(chan struct{}),
		events: make(chan event, runtime.NumCPU()),

		outputBuffer: bytes.NewBuffer(nil),
		dateBuffer:   make([]byte, 25),

		level:     levelInfo,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}

	go logger.Start()
}

// New creates a Logger
func New(config Config) Logger {
	level, err := parseLevel(strings.TrimSpace(*config.level))

	logger := Logger{
		done:   make(chan struct{}),
		events: make(chan event, runtime.NumCPU()),

		outputBuffer: bytes.NewBuffer(nil),
		dateBuffer:   make([]byte, 25),

		level:     level,
		outWriter: os.Stdout,
		errWriter: os.Stderr,

		jsonFormat: *config.json,
		timeKey:    EscapeString(strings.TrimSpace(*config.timeKey)),
		levelKey:   EscapeString(strings.TrimSpace(*config.levelKey)),
		messageKey: EscapeString(strings.TrimSpace(*config.messageKey)),
	}

	go logger.Start()

	if err != nil {
		logger.Error(err.Error())
	}

	return logger
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
			safeErrorWrite(fmt.Sprintf("unable to write log: %s\n", err))
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
	l.output(levelDebug, format, a...)
}

// Debug logs debug message
func (l Logger) Debug(format string, a ...interface{}) {
	l.output(levelDebug, format, a...)
}

// Info logs info message
func (l Logger) Info(format string, a ...interface{}) {
	l.output(levelInfo, format, a...)
}

// Warn logs warning message
func (l Logger) Warn(format string, a ...interface{}) {
	l.output(levelWarning, format, a...)
}

// Error logs error message
func (l Logger) Error(format string, a ...interface{}) {
	l.output(levelError, format, a...)
}

// Fatal logs error message and exit with status code 1
func (l Logger) Fatal(err error) {
	if err == nil {
		return
	}

	l.output(levelFatal, "%s", err)
	l.Close()

	exitFunc(1)
}

func (l Logger) output(lev level, format string, a ...interface{}) {
	if l.level < lev {
		return
	}

	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}

	l.events <- event{timestamp: nowFunc(), level: lev, message: message}
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
	l.outputBuffer.WriteString(EscapeString(e.message))
	l.outputBuffer.WriteString(`"}`)
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
	l.outputBuffer.WriteString("\n")

	return l.outputBuffer.Bytes()
}

func safeErrorWrite(message string) {
	if _, err := os.Stderr.WriteString(message); err != nil {
		// do nothing here
	}
}

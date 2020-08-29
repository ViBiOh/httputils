package logger

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
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
	levelValues = []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

	logger   *Logger
	exitFunc = os.Exit
)

func parseLevel(line string) (level, error) {
	for i, l := range levelValues {
		if strings.EqualFold(l, line) {
			return level(i), nil
		}
	}

	return levelInfo, fmt.Errorf("invalid value `%s` for level", line)
}

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
	builder bytes.Buffer
	buffer  chan event
	wg      sync.WaitGroup

	json       bool
	timeKey    string
	levelKey   string
	messageKey string

	level     level
	outWriter io.Writer
	errWriter io.Writer
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		level:      flags.New(prefix, "logger").Name("Level").Default("INFO").Label("Logger level").ToString(fs),
		json:       flags.New(prefix, "logger").Name("Json").Default(false).Label("Log format as JSON").ToBool(fs),
		timeKey:    flags.New(prefix, "logger").Name("TimeKey").Default("time").Label("Key for timestam in JSON").ToString(fs),
		levelKey:   flags.New(prefix, "logger").Name("LevelKey").Default("level").Label("Key for level in JSON").ToString(fs),
		messageKey: flags.New(prefix, "logger").Name("MessageKey").Default("message").Label("Key for message in JSON").ToString(fs),
	}
}

func init() {
	logger = &Logger{
		buffer: make(chan event, runtime.NumCPU()),

		level:     levelInfo,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}

	go logger.Start()
}

// New creates a Logger
func New(config Config) *Logger {
	level, err := parseLevel(*config.level)

	logger := Logger{
		buffer: make(chan event, runtime.NumCPU()),

		level:     level,
		outWriter: os.Stdout,
		errWriter: os.Stderr,

		json:       *config.json,
		timeKey:    EscapeString(*config.timeKey),
		levelKey:   EscapeString(*config.levelKey),
		messageKey: EscapeString(*config.messageKey),
	}

	go logger.Start()

	if err != nil {
		logger.Error(err.Error())
	}

	return &logger
}

// Start starts logger's writer
func (l *Logger) Start() {
	l.wg.Add(1)
	defer l.wg.Done()

	var payload []byte
	var err error

	for e := range l.buffer {
		if l.json {
			payload = e.json(l)
		} else {
			payload = e.text(l)
		}

		if e.level <= levelInfo {
			_, err = l.outWriter.Write(payload)
		} else {
			_, err = l.errWriter.Write(payload)
		}

		if err != nil {
			writeError(fmt.Sprintf("unable to write log: %s\n", err))
		}
	}
}

// Close ends logger gracefully
func (l *Logger) Close() {
	close(l.buffer)
	l.wg.Wait()
}

// Trace logs tracing message
func (l *Logger) Trace(format string, a ...interface{}) {
	l.output(levelDebug, format, a...)
}

// Debug logs debug message
func (l *Logger) Debug(format string, a ...interface{}) {
	l.output(levelDebug, format, a...)
}

// Info logs info message
func (l *Logger) Info(format string, a ...interface{}) {
	l.output(levelInfo, format, a...)
}

// Warn logs warning message
func (l *Logger) Warn(format string, a ...interface{}) {
	l.output(levelWarning, format, a...)
}

// Error logs error message
func (l *Logger) Error(format string, a ...interface{}) {
	l.output(levelError, format, a...)
}

// Fatal logs error message and exit with status code 1
func (l *Logger) Fatal(err error) {
	if err == nil {
		return
	}

	l.output(levelFatal, "%s", err)
	l.Close()

	if closer, ok := l.outWriter.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			writeError(fmt.Sprintf("unable to close out writer: %s\n", err))
		}
	}

	if closer, ok := l.errWriter.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			writeError(fmt.Sprintf("unable to close err writer: %s\n", err))
		}
	}

	exitFunc(1)
}

func (l *Logger) output(lev level, format string, a ...interface{}) {
	if l.level < lev {
		return
	}

	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}

	l.buffer <- event{time.Now(), lev, message}
}

func writeError(message string) {
	if _, err := os.Stderr.WriteString(message); err != nil {
		// do nothing here
	}
}

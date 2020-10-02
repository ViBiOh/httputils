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

var (
	logger     *Logger
	exitFunc   = os.Exit
	nowFunc    = time.Now
	dateBuffer = make([]byte, 25)
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
	logBuffer  bytes.Buffer
	eventsChan chan event
	wg         sync.WaitGroup

	jsonFormat bool
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
		timeKey:    flags.New(prefix, "logger").Name("TimeKey").Default("time").Label("Key for timestamp in JSON").ToString(fs),
		levelKey:   flags.New(prefix, "logger").Name("LevelKey").Default("level").Label("Key for level in JSON").ToString(fs),
		messageKey: flags.New(prefix, "logger").Name("MessageKey").Default("message").Label("Key for message in JSON").ToString(fs),
	}
}

func init() {
	logger = &Logger{
		eventsChan: make(chan event, runtime.NumCPU()),

		level:     levelInfo,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}

	logger.wg.Add(1)
	go logger.Start()
}

// New creates a Logger
func New(config Config) *Logger {
	level, err := parseLevel(strings.TrimSpace(*config.level))

	logger := Logger{
		eventsChan: make(chan event, runtime.NumCPU()),

		level:     level,
		outWriter: os.Stdout,
		errWriter: os.Stderr,

		jsonFormat: *config.json,
		timeKey:    EscapeString(strings.TrimSpace(*config.timeKey)),
		levelKey:   EscapeString(strings.TrimSpace(*config.levelKey)),
		messageKey: EscapeString(strings.TrimSpace(*config.messageKey)),
	}

	logger.wg.Add(1)
	go logger.Start()

	if err != nil {
		logger.Error(err.Error())
	}

	return &logger
}

// Start starts logger's writer
func (l *Logger) Start() {
	defer l.wg.Done()

	var payload []byte
	var err error

	for e := range l.eventsChan {
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
}

// Close ends logger gracefully
func (l *Logger) Close() {
	close(l.eventsChan)
	l.wg.Wait()

	if l.outWriter != os.Stdout {
		if closer, ok := l.outWriter.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				safeErrorWrite(fmt.Sprintf("unable to close out writer: %s\n", err))
			}
		}
	}

	if l.outWriter != os.Stderr {
		if closer, ok := l.errWriter.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				safeErrorWrite(fmt.Sprintf("unable to close err writer: %s\n", err))
			}
		}
	}
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

	l.eventsChan <- event{nowFunc(), lev, message}
}

func (l *Logger) json(e event) []byte {
	l.logBuffer.Reset()

	dateBuffer = e.timestamp.AppendFormat(dateBuffer[:0], time.RFC3339)

	l.logBuffer.WriteString(`{"`)
	l.logBuffer.WriteString(l.timeKey)
	l.logBuffer.WriteString(`":"`)
	l.logBuffer.Write(dateBuffer)
	l.logBuffer.WriteString(`","`)
	l.logBuffer.WriteString(l.levelKey)
	l.logBuffer.WriteString(`":"`)
	l.logBuffer.WriteString(levelValues[e.level])
	l.logBuffer.WriteString(`","`)
	l.logBuffer.WriteString(l.messageKey)
	l.logBuffer.WriteString(`":"`)
	l.logBuffer.WriteString(EscapeString(e.message))
	l.logBuffer.WriteString(`"}`)
	l.logBuffer.WriteString("\n")

	return l.logBuffer.Bytes()
}

func (l *Logger) text(e event) []byte {
	l.logBuffer.Reset()

	dateBuffer = e.timestamp.AppendFormat(dateBuffer[:0], time.RFC3339)

	l.logBuffer.Write(dateBuffer)
	l.logBuffer.WriteString(` `)
	l.logBuffer.WriteString(levelValues[e.level])
	l.logBuffer.WriteString(` `)
	l.logBuffer.WriteString(e.message)
	l.logBuffer.WriteString("\n")

	return l.logBuffer.Bytes()
}

func safeErrorWrite(message string) {
	if _, err := os.Stderr.WriteString(message); err != nil {
		// do nothing here
	}
}

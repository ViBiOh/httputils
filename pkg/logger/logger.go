package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	logger = New(false)
}

// Global sets global logger
func Global(l *Logger) {
	logger.Close()
	logger = l
}

// Logger defines a logger instance
type Logger struct {
	builder bytes.Buffer
	buffer  chan event
	wg      sync.WaitGroup
	json    bool

	outWriter io.Writer
	errWriter io.Writer
}

// New creates a Logger
func New(json bool) *Logger {
	logger := Logger{
		buffer:    make(chan event, runtime.NumCPU()),
		outWriter: os.Stdout,
		errWriter: os.Stderr,
		json:      json,
	}

	go logger.Start()

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
			payload = e.json(&l.builder)
		} else {
			payload = e.text(&l.builder)
		}

		if e.level <= levelInfo {
			_, err = l.outWriter.Write(payload)
		} else {
			_, err = l.errWriter.Write(payload)
		}

		if err != nil {
			log.Printf("unable to write log: %s", err)
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
	exitFunc(1)
}

func (l *Logger) output(lev level, format string, a ...interface{}) {
	msg := format
	if len(a) > 0 {
		msg = fmt.Sprintf(format, a...)
	}

	l.buffer <- event{time.Now().Unix(), lev, msg}
}

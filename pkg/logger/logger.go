package logger

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ViBiOh/httputils/pkg/errors"
)

// LogReporter describes a log reporter
type LogReporter interface {
	Info(string)
	Warn(string)
	Error(string)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var (
	info = log.New(os.Stdout, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	erro = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile|log.LUTC)

	reporters = make([]LogReporter, 0)
	mutex     = sync.RWMutex{}
)

// AddReporter add a LogReporter to the list
func AddReporter(reporter LogReporter) {
	mutex.Lock()
	defer mutex.Unlock()

	reporters = append(reporters, reporter)
}

func output(l *log.Logger, format string, a ...interface{}) string {
	content := fmt.Sprintf(format, a...)
	if err := l.Output(3, content); err != nil {
		log.Printf("%#v", errors.WithStack(err))
	}

	return content
}

// Info log info message
func Info(format string, a ...interface{}) {
	content := output(info, format, a...)

	for _, reporter := range reporters {
		reporter.Info(content)
	}
}

// Warn log warning message
func Warn(format string, a ...interface{}) {
	content := output(warn, format, a...)

	for _, reporter := range reporters {
		reporter.Warn(content)
	}
}

// Error log error message
func Error(format string, a ...interface{}) {
	content := output(erro, format, a...)

	for _, reporter := range reporters {
		reporter.Error(content)
	}
}

// Fatal log error message and exit
func Fatal(format string, a ...interface{}) {
	content := output(erro, format, a...)

	for _, reporter := range reporters {
		reporter.Error(content)
	}

	os.Exit(1)
}

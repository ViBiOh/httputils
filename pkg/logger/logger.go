package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	callDepth = 3
)

var (
	exitFunc = os.Exit

	info = log.New(os.Stdout, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	erro = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile|log.LUTC)
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

// Info logs info message
func Info(format string, a ...interface{}) {
	output(info, format, a...)
}

// Warn logs warning message
func Warn(format string, a ...interface{}) {
	output(warn, format, a...)
}

// Error logs error message
func Error(format string, a ...interface{}) {
	output(erro, format, a...)
}

// Fatal logs error message and exit with status code 1
func Fatal(err error) {
	if err == nil {
		return
	}

	output(erro, "%s", err)
	exitFunc(1)
}

func output(l *log.Logger, format string, a ...interface{}) {
	if err := l.Output(callDepth, fmt.Sprintf(format, a...)); err != nil {
		log.Printf("%s", err)
	}
}

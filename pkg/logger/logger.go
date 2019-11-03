package logger

import (
	"fmt"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var (
	info = log.New(os.Stdout, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	erro = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile|log.LUTC)
)

func output(l *log.Logger, format string, a ...interface{}) {
	if err := l.Output(3, fmt.Sprintf(format, a...)); err != nil {
		log.Printf("%s", err)
	}
}

// Info log info message
func Info(format string, a ...interface{}) {
	output(info, format, a...)
}

// Warn log warning message
func Warn(format string, a ...interface{}) {
	output(warn, format, a...)
}

// Error log error message
func Error(format string, a ...interface{}) {
	output(erro, format, a...)
}

// Fatal log error message and exit
func Fatal(err error) {
	if err == nil {
		return
	}

	output(erro, "%s", err)
	os.Exit(1)
}

package logger

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var (
	info = log.New(os.Stdout, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	erro = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile|log.LUTC)

	mutex = sync.RWMutex{}
)

func output(l *log.Logger, format string, a ...interface{}) string {
	content := fmt.Sprintf(format, a...)
	if err := l.Output(3, content); err != nil {
		log.Printf("%#v", errors.WithStack(err))
	}

	return content
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

	output(erro, "%#v", err)
	os.Exit(1)
}

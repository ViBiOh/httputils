package logger

import (
	"io/ioutil"
	"runtime"
	"testing"
	"time"
)

func BenchmarkJson(b *testing.B) {
	logger := Logger{
		json:      true,
		outWriter: ioutil.Discard,
		level:     levelInfo,
	}

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
	}

	for i := 0; i < b.N; i++ {
		e.json(&logger)
	}
}

func BenchmarkText(b *testing.B) {
	logger := Logger{
		outWriter: ioutil.Discard,
		level:     levelInfo,
	}

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
	}

	for i := 0; i < b.N; i++ {
		e.text(&logger)
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	logger := Logger{
		outWriter: ioutil.Discard,
		level:     levelInfo,
		buffer:    make(chan event, runtime.NumCPU()),
	}

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.output(levelInfo, "Hello world")
	}
}

func BenchmarkFormattedOutput(b *testing.B) {
	logger := Logger{
		outWriter: ioutil.Discard,
		level:     levelInfo,
		buffer:    make(chan event, runtime.NumCPU()),
	}

	go logger.Start()
	defer logger.Close()

	time := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		logger.output(levelInfo, "Hello %s, it's %d", "Bob", time)
	}
}

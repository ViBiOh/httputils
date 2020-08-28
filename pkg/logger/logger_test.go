package logger

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"
)

func BenchmarkJson(b *testing.B) {
	e := event{
		ts:    time.Now().Unix(),
		level: levelInfo,
		msg:   "Hello world",
	}

	builder := bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		e.json(&builder)
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	l := New(true)
	l.outWriter = ioutil.Discard

	for i := 0; i < b.N; i++ {
		l.output(levelInfo, "Hello world")
	}
}

func BenchmarkFormattedOutput(b *testing.B) {
	l := New(true)
	l.outWriter = ioutil.Discard
	time := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		l.output(levelInfo, "Hello %s, it's %d", "Bob", time)
	}
}

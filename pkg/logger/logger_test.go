package logger

import (
	"io/ioutil"
	"runtime"
	"testing"
	"time"
)

func TestText(t *testing.T) {
	type args struct {
		e event
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"simple",
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			"2020-09-30T14:59:38Z INFO Hello world\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			logger := Logger{}

			if got := logger.text(tc.args.e); string(got) != tc.want {
				t.Errorf("text() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestJson(t *testing.T) {
	type args struct {
		e event
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"simple",
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			`{"ts":"2020-09-30T14:59:38Z","level":"INFO","msg":"Hello world"}
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			logger := Logger{
				timeKey:    "ts",
				levelKey:   "level",
				messageKey: "msg",
			}

			if got := logger.json(tc.args.e); string(got) != tc.want {
				t.Errorf("json() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkJson(b *testing.B) {
	logger := Logger{
		jsonFormat: true,
		outWriter:  ioutil.Discard,
		level:      levelInfo,
	}

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
	}

	for i := 0; i < b.N; i++ {
		logger.json(e)
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
		logger.text(e)
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	logger := Logger{
		outWriter: ioutil.Discard,
		level:     levelInfo,
		buffer:    make(chan event, runtime.NumCPU()),
	}

	logger.wg.Add(1)
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

	logger.wg.Add(1)
	go logger.Start()
	defer logger.Close()

	time := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		logger.output(levelInfo, "Hello %s, it's %d", "Bob", time)
	}
}

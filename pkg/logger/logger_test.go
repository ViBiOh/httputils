package logger

import (
	"bytes"
	"flag"
	"io/ioutil"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -json\n    \t[logger] Log format as JSON {SIMPLE_JSON}\n  -level string\n    \t[logger] Logger level {SIMPLE_LEVEL} (default \"INFO\")\n  -levelKey string\n    \t[logger] Key for level in JSON {SIMPLE_LEVEL_KEY} (default \"level\")\n  -messageKey string\n    \t[logger] Key for message in JSON {SIMPLE_MESSAGE_KEY} (default \"message\")\n  -timeKey string\n    \t[logger] Key for timestamp in JSON {SIMPLE_TIME_KEY} (default \"time\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestStart(t *testing.T) {
	type args struct {
		e event
	}

	var cases = []struct {
		intention string
		instance  *Logger
		args      args
		want      string
	}{
		{
			"json",
			&Logger{
				buffer:     make(chan event, runtime.NumCPU()),
				jsonFormat: true,
				timeKey:    "ts",
				levelKey:   "level",
				messageKey: "msg",
				level:      levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			`{"ts":"2020-09-30T14:59:38Z","level":"INFO","msg":"Hello world"}
`,
		},
		{
			"text",
			&Logger{
				buffer: make(chan event, runtime.NumCPU()),
				level:  levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			"2020-09-30T14:59:38Z INFO Hello world\n",
		},
		{
			"error",
			&Logger{
				buffer: make(chan event, runtime.NumCPU()),
				level:  levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelDebug, message: "Hello world"},
			},
			"2020-09-30T14:59:38Z DEBUG Hello world\n",
		},
	}

	for _, tc := range cases {
		outputter := bytes.Buffer{}
		tc.instance.outWriter = &outputter
		tc.instance.errWriter = &outputter
		tc.instance.wg.Add(1)
		go tc.instance.Start()

		t.Run(tc.intention, func(t *testing.T) {
			tc.instance.buffer <- tc.args.e
			tc.instance.Close()

			if got := outputter.String(); got != tc.want {
				t.Errorf("Start() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestOutput(t *testing.T) {
	type args struct {
		lev    level
		format string
		a      []interface{}
	}

	now = func() time.Time {
		return time.Date(2020, 9, 21, 18, 34, 57, 0, time.UTC)
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"ignored",
			args{
				lev:    levelDebug,
				format: "Hello World",
			},
			"",
		},
		{
			"info",
			args{
				lev:    levelInfo,
				format: "Hello World",
			},
			"2020-09-21T18:34:57Z INFO Hello World\n",
		},
		{
			"format",
			args{
				lev:    levelInfo,
				format: "Hello %s",
				a:      []interface{}{"World"},
			},
			"2020-09-21T18:34:57Z INFO Hello World\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := bytes.Buffer{}
			logger := Logger{
				outWriter: &writer,
				level:     levelInfo,
				buffer:    make(chan event, runtime.NumCPU()),
			}

			logger.wg.Add(1)
			go logger.Start()

			logger.output(tc.args.lev, tc.args.format, tc.args.a...)
			logger.Close()

			got := writer.String()

			if got != tc.want {
				t.Errorf("output() = `%s`, want `%s`", got, tc.want)
			}
		})
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

func BenchmarkNoOutput(b *testing.B) {
	logger := Logger{
		outWriter: ioutil.Discard,
		level:     levelWarning,
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

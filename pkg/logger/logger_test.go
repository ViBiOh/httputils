package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"testing"
	"time"
)

type writeCloser struct {
	error
	io.Writer
}

func (wc writeCloser) Close() error {
	return wc.error
}

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -json\n    \t[logger] Log format as JSON {SIMPLE_JSON}\n  -level string\n    \t[logger] Logger level {SIMPLE_LEVEL} (default \"INFO\")\n  -levelKey string\n    \t[logger] Key for level in JSON {SIMPLE_LEVEL_KEY} (default \"level\")\n  -messageKey string\n    \t[logger] Key for message in JSON {SIMPLE_MESSAGE_KEY} (default \"message\")\n  -timeKey string\n    \t[logger] Key for timestamp in JSON {SIMPLE_TIME_KEY} (default \"time\")\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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
	t.Parallel()

	type args struct {
		e event
	}

	cases := map[string]struct {
		instance Logger
		args     args
		want     string
	}{
		"json": {
			Logger{
				done:         make(chan struct{}),
				events:       make(chan event, runtime.NumCPU()),
				outputBuffer: bytes.NewBuffer(nil),
				jsonFormat:   true,
				timeKey:      "ts",
				levelKey:     "level",
				messageKey:   "msg",
				level:        levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			`{"ts":"2020-09-30T14:59:38Z","level":"INFO","msg":"Hello world"}
`,
		},
		"text": {
			Logger{
				done:         make(chan struct{}),
				events:       make(chan event, runtime.NumCPU()),
				outputBuffer: bytes.NewBuffer(nil),
				level:        levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelInfo, message: "Hello world"},
			},
			"2020-09-30T14:59:38Z INFO Hello world\n",
		},
		"error": {
			Logger{
				done:         make(chan struct{}),
				events:       make(chan event, runtime.NumCPU()),
				outputBuffer: bytes.NewBuffer(nil),
				level:        levelInfo,
			},
			args{
				e: event{timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC), level: levelDebug, message: "Hello world"},
			},
			"2020-09-30T14:59:38Z DEBUG Hello world\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		writer := bytes.NewBuffer(nil)
		testCase.instance.outWriter = writer
		testCase.instance.errWriter = writer
		go testCase.instance.Start()

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			testCase.instance.events <- testCase.args.e
			testCase.instance.Close()

			if got := writer.String(); got != testCase.want {
				t.Errorf("Start() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	type args struct {
		out io.Writer
		err io.Writer
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"simple": {
			args{
				out: io.Discard,
			},
			true,
		},
		"closer": {
			args{
				out: writeCloser{nil, bytes.NewBuffer(nil)},
				err: writeCloser{nil, bytes.NewBuffer(nil)},
			},
			true,
		},
		"closer error": {
			args{
				out: writeCloser{errors.New("error"), bytes.NewBuffer(nil)},
				err: writeCloser{errors.New("error"), bytes.NewBuffer(nil)},
			},
			true,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			logger := newLogger(testCase.args.out, testCase.args.err, levelInfo, false, "time", "level", "msg")

			go logger.Start()

			logger.Debug("Hello World")
			logger.Trace("Hello World")
			logger.Info("Hello World")
			logger.Warn("Hello World")
			logger.Error("Hello World")
			logger.Close()
		})
	}
}

func TestOutput(t *testing.T) {
	t.Parallel()

	type args struct {
		lev    level
		format string
		a      []any
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"ignored": {
			args{
				lev:    levelDebug,
				format: "Hello World",
			},
			"",
		},
		"info": {
			args{
				lev:    levelInfo,
				format: "Hello World",
			},
			"2020-09-21T18:34:57Z INFO Hello World\n",
		},
		"format": {
			args{
				lev:    levelInfo,
				format: "Hello %s",
				a:      []any{"World"},
			},
			"2020-09-21T18:34:57Z INFO Hello World\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := bytes.NewBuffer(nil)
			logger := newLogger(writer, writer, levelInfo, false, "time", "level", "msg")
			logger.clock = func() time.Time { return time.Date(2020, 9, 21, 18, 34, 57, 0, time.UTC) }

			go logger.Start()

			logger.output(testCase.args.lev, nil, testCase.args.format, testCase.args.a...)
			logger.Close()

			got := writer.String()

			if got != testCase.want {
				t.Errorf("output() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func BenchmarkStandardSimpleOutput(b *testing.B) {
	logger := log.New(io.Discard, "", log.Ldate|log.Ltime)

	for i := 0; i < b.N; i++ {
		logger.Print("[INFO] Hello world")
	}
}

func BenchmarkStandardSimpleFormattedOutput(b *testing.B) {
	logger := log.New(io.Discard, "", log.Ldate|log.Ltime)
	now := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		logger.Printf("Hello %s, it's %d", "Bob", now)
	}
}

func BenchmarkNoOutput(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelWarning, false, "time", "level", "msg")
	logger.clock = time.Now

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello world")
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = time.Now

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello world")
	}
}

func BenchmarkFormattedOutput(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = time.Now

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello %s", "Bob")
	}
}

func BenchmarkFormattedOutputFields(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = time.Now

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.WithField("success", true).WithField("count", 7).Info("Hello %s", "Bob")
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	type args struct {
		e event
	}

	cases := map[string]struct {
		args args
		want map[string]any
	}{
		"simple": {
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
				},
			},
			map[string]any{
				"ts":    "2020-09-30T14:59:38Z",
				"level": "INFO",
				"msg":   "Hello world",
			},
		},
		"with fields": {
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: []field{
						{
							name:  "count",
							value: 7,
						}, {
							name:  "name",
							value: "test",
						}, {
							name:  "success",
							value: true,
						},
					},
				},
			},
			map[string]any{
				"count":   7,
				"level":   "INFO",
				"msg":     "Hello world",
				"name":    "test",
				"success": true,
				"ts":      "2020-09-30T14:59:38Z",
			},
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			logger := Logger{
				outputBuffer: bytes.NewBuffer(nil),
				timeKey:      "ts",
				levelKey:     "level",
				messageKey:   "msg",
			}

			var values map[string]any
			if err := json.Unmarshal(logger.json(testCase.args.e), &values); err != nil {
				t.Errorf("unmarshal json payload: %s", err)
			}

			if fmt.Sprintf("%+v", values) != fmt.Sprintf("%+v", testCase.want) {
				t.Errorf("json() = %+v, want %+v", values, testCase.want)
			}
		})
	}
}

func BenchmarkJSON(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, true, "time", "level", "msg")

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
	}

	for i := 0; i < b.N; i++ {
		logger.json(e)
	}
}

func BenchmarkJSONWithFields(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, true, "time", "level", "msg")

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
		fields: []field{
			{
				name:  "count",
				value: 7,
			},
			{
				name:  "name",
				value: "test",
			},
			{
				name:  "success",
				value: true,
			},
		},
	}

	for i := 0; i < b.N; i++ {
		logger.json(e)
	}
}

func TestText(t *testing.T) {
	t.Parallel()

	type args struct {
		e event
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"simple": {
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world\n",
		},
		"string field": {
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: []field{
						{
							name:  "name",
							value: "test",
						},
					},
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world name=\"test\"\n",
		},
		"numeric field": {
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: []field{
						{
							name:  "count",
							value: 7,
						},
					},
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world count=7\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			logger := Logger{
				outputBuffer: bytes.NewBuffer(nil),
			}

			if got := logger.text(testCase.args.e); !strings.Contains(string(got), testCase.want) {
				t.Errorf("text() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func BenchmarkText(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
	}

	for i := 0; i < b.N; i++ {
		logger.text(e)
	}
}

func BenchmarkTextWithFields(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")

	e := event{
		timestamp: time.Now(),
		level:     levelInfo,
		message:   "Hello world",
		fields: []field{
			{
				name:  "count",
				value: 7,
			},
			{
				name:  "name",
				value: "test",
			},
			{
				name:  "success",
				value: true,
			},
		},
	}

	for i := 0; i < b.N; i++ {
		logger.text(e)
	}
}

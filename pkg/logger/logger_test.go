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

	"github.com/ViBiOh/httputils/v4/pkg/clock"
)

type writeCloser struct {
	error
	io.Writer
}

func (wc writeCloser) Close() error {
	return wc.error
}

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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
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
		instance  Logger
		args      args
		want      string
	}{
		{
			"json",
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
		{
			"text",
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
		{
			"error",
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

	for _, tc := range cases {
		writer := bytes.NewBuffer(nil)
		tc.instance.outWriter = writer
		tc.instance.errWriter = writer
		go tc.instance.Start()

		t.Run(tc.intention, func(t *testing.T) {
			tc.instance.events <- tc.args.e
			tc.instance.Close()

			if got := writer.String(); got != tc.want {
				t.Errorf("Start() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	type args struct {
		out io.Writer
		err io.Writer
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
	}{
		{
			"simple",
			args{
				out: io.Discard,
			},
			true,
		},
		{
			"closer",
			args{
				out: writeCloser{nil, bytes.NewBuffer(nil)},
				err: writeCloser{nil, bytes.NewBuffer(nil)},
			},
			true,
		},
		{
			"closer error",
			args{
				out: writeCloser{errors.New("error"), bytes.NewBuffer(nil)},
				err: writeCloser{errors.New("error"), bytes.NewBuffer(nil)},
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			logger := newLogger(tc.args.out, tc.args.err, levelInfo, false, "time", "level", "msg")

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
	type args struct {
		lev    level
		format string
		a      []interface{}
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
			writer := bytes.NewBuffer(nil)
			logger := newLogger(writer, writer, levelInfo, false, "time", "level", "msg")
			logger.clock = clock.New(time.Date(2020, 9, 21, 18, 34, 57, 0, time.UTC))

			go logger.Start()

			logger.output(tc.args.lev, nil, tc.args.format, tc.args.a...)
			logger.Close()

			got := writer.String()

			if got != tc.want {
				t.Errorf("output() = `%s`, want `%s`", got, tc.want)
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
	logger.clock = clock.New(time.Now())

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello world")
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = clock.New(time.Now())

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello world")
	}
}

func BenchmarkFormattedOutput(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = clock.New(time.Now())

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.Info("Hello %s", "Bob")
	}
}

func BenchmarkFormattedOutputFields(b *testing.B) {
	logger := newLogger(io.Discard, io.Discard, levelInfo, false, "time", "level", "msg")
	logger.clock = clock.New(time.Now())

	go logger.Start()
	defer logger.Close()

	for i := 0; i < b.N; i++ {
		logger.WithField("success", true).WithField("count", 7).Info("Hello %s", "Bob")
	}
}

func TestJSON(t *testing.T) {
	type args struct {
		e event
	}

	var cases = []struct {
		intention string
		args      args
		want      map[string]interface{}
	}{
		{
			"simple",
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
				},
			},
			map[string]interface{}{
				"ts":    "2020-09-30T14:59:38Z",
				"level": "INFO",
				"msg":   "Hello world",
			},
		},
		{
			"with fields",
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: map[string]interface{}{
						"count":   7,
						"name":    "test",
						"success": true,
					},
				},
			},
			map[string]interface{}{
				"count":   7,
				"level":   "INFO",
				"msg":     "Hello world",
				"name":    "test",
				"success": true,
				"ts":      "2020-09-30T14:59:38Z",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			logger := Logger{
				outputBuffer: bytes.NewBuffer(nil),
				timeKey:      "ts",
				levelKey:     "level",
				messageKey:   "msg",
			}

			var values map[string]interface{}
			if err := json.Unmarshal([]byte(logger.json(tc.args.e)), &values); err != nil {
				t.Errorf("unable to unmarshal json payload: %s", err)
			}

			if fmt.Sprintf("%+v", values) != fmt.Sprintf("%+v", tc.want) {
				t.Errorf("json() = %+v, want %+v", values, tc.want)
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
		fields: map[string]interface{}{
			"count":   7,
			"name":    "test",
			"success": true,
		},
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
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world\n",
		},
		{
			"string field",
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: map[string]interface{}{
						"name": "test",
					},
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world name=\"test\"\n",
		},
		{
			"numeric field",
			args{
				e: event{
					timestamp: time.Date(2020, 9, 30, 14, 59, 38, 0, time.UTC),
					level:     levelInfo,
					message:   "Hello world",
					fields: map[string]interface{}{
						"count": 7,
					},
				},
			},
			"2020-09-30T14:59:38Z INFO Hello world count=7\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			logger := Logger{
				outputBuffer: bytes.NewBuffer(nil),
			}

			if got := logger.text(tc.args.e); !strings.Contains(string(got), tc.want) {
				t.Errorf("text() = `%s`, want `%s`", got, tc.want)
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
		fields: map[string]interface{}{
			"count":   7,
			"name":    "test",
			"success": true,
		},
	}

	for i := 0; i < b.N; i++ {
		logger.text(e)
	}
}

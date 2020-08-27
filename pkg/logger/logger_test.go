package logger

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"
)

func TestBasics(t *testing.T) {
	var cases = []struct {
		intention string
		method    func(string, ...interface{})
		want      string
	}{
		{
			"info",
			Info,
			"TEST  Success\n",
		},
		{
			"warn",
			Warn,
			"TEST  Success\n",
		},
		{
			"error",
			Error,
			"TEST  Success\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := &strings.Builder{}
			info = log.New(writer, "TEST  ", 0)
			warn = log.New(writer, "TEST  ", 0)
			erro = log.New(writer, "TEST  ", 0)

			testCase.method("Success")

			if result := writer.String(); result != testCase.want {
				t.Errorf("`%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestFatal(t *testing.T) {
	var cases = []struct {
		intention string
		err       error
		want      int
		wantLog   string
	}{
		{
			"no error",
			nil,
			0,
			"",
		},
		{
			"error and exit",
			errors.New("testing logger"),
			1,
			"TEST  testing logger\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			var result int
			exitFunc = func(status int) {
				result = status
			}

			writer := &strings.Builder{}
			erro = log.New(writer, "TEST  ", 0)

			Fatal(testCase.err)

			if result != testCase.want {
				t.Errorf("Fatal() = %d, want %d", result, testCase.want)
			}

			if result := writer.String(); result != testCase.wantLog {
				t.Errorf("Fatal() = `%s`, want `%s`", result, testCase.wantLog)
			}
		})
	}
}

func TestOutput(t *testing.T) {
	var cases = []struct {
		intention string
		format    string
		params    []interface{}
		want      string
	}{
		{
			"empty",
			"",
			nil,
			"TEST  \n",
		},
		{
			"format",
			"%s: %d run, result=%t",
			[]interface{}{"formatting test", 2, true},
			"TEST  formatting test: 2 run, result=true\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := &strings.Builder{}
			logger := log.New(writer, "TEST  ", 0)

			output(logger, testCase.format, testCase.params...)

			if result := writer.String(); result != testCase.want {
				t.Errorf("Info() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func BenchmarkSimpleOutput(b *testing.B) {
	logger := log.New(ioutil.Discard, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)

	for i := 0; i < b.N; i++ {
		output(logger, "Hello world")
	}
}

func BenchmarkSimpleOutputWithoutFile(b *testing.B) {
	logger := log.New(ioutil.Discard, "INFO  ", log.LstdFlags|log.LUTC)

	for i := 0; i < b.N; i++ {
		output(logger, "Hello world")
	}
}

func BenchmarkFormattedOutput(b *testing.B) {
	logger := log.New(ioutil.Discard, "INFO  ", log.LstdFlags|log.Lshortfile|log.LUTC)
	time := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		output(logger, "Hello %s, it's %d", "Bob", time)
	}
}

func BenchmarkFormattedOutputWithoutFile(b *testing.B) {
	logger := log.New(ioutil.Discard, "INFO  ", log.LstdFlags|log.LUTC)
	time := time.Now().Unix()

	for i := 0; i < b.N; i++ {
		output(logger, "Hello %s, it's %d", "Bob", time)
	}
}

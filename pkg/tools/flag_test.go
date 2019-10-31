package tools

import (
	"flag"
	"reflect"
	"strings"
	"testing"
)

func TestNewFlag(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		docPrefix string
		want      *Flag
	}{
		{
			"simple",
			"newFlag",
			"test",
			&Flag{
				prefix:    "NewFlag",
				docPrefix: "newFlag",
			},
		},
		{
			"without prefix",
			"",
			"test",
			&Flag{
				prefix:    "",
				docPrefix: "test",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := NewFlag(testCase.prefix, testCase.docPrefix); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("NewFlag() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	var cases = []struct {
		intention    string
		prefix       string
		docPrefix    string
		name         string
		defaultValue string
		label        string
		want         string
	}{
		{
			"simple",
			"",
			"cli",
			"test",
			"",
			"Test flag",
			"Usage of ToString:\n  -test string\n    \t[cli] Test flag {TO_STRING_TEST}\n",
		},
		{
			"with prefix",
			"context",
			"cli",
			"test",
			"default",
			"Test flag",
			"Usage of ToString:\n  -contextTest string\n    \t[context] Test flag {TO_STRING_CONTEXT_TEST} (default \"default\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToString", flag.ContinueOnError)
			fg := NewFlag(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToString(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToString() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	var cases = []struct {
		intention    string
		prefix       string
		docPrefix    string
		name         string
		defaultValue int
		label        string
		want         string
	}{
		{
			"simple",
			"",
			"cli",
			"test",
			0,
			"Test flag",
			"Usage of ToInt:\n  -test int\n    \t[cli] Test flag {TO_INT_TEST}\n",
		},
		{
			"with prefix",
			"context",
			"cli",
			"test",
			8000,
			"Test flag",
			"Usage of ToInt:\n  -contextTest int\n    \t[context] Test flag {TO_INT_CONTEXT_TEST} (default 8000)\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToInt", flag.ContinueOnError)
			fg := NewFlag(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToInt(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToInt() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestToUint(t *testing.T) {
	var cases = []struct {
		intention    string
		prefix       string
		docPrefix    string
		name         string
		defaultValue uint
		label        string
		want         string
	}{
		{
			"simple",
			"",
			"cli",
			"test",
			0,
			"Test flag",
			"Usage of ToUint:\n  -test uint\n    \t[cli] Test flag {TO_UINT_TEST}\n",
		},
		{
			"with prefix",
			"context",
			"cli",
			"test",
			8000,
			"Test flag",
			"Usage of ToUint:\n  -contextTest uint\n    \t[context] Test flag {TO_UINT_CONTEXT_TEST} (default 8000)\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToUint", flag.ContinueOnError)
			fg := NewFlag(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToUint(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToUint() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	var cases = []struct {
		intention    string
		prefix       string
		docPrefix    string
		name         string
		defaultValue float64
		label        string
		want         string
	}{
		{
			"simple",
			"",
			"cli",
			"test",
			0,
			"Test flag",
			"Usage of ToFloat64:\n  -test float\n    \t[cli] Test flag {TO_FLOAT64_TEST}\n",
		},
		{
			"with prefix",
			"context",
			"cli",
			"test",
			12.34,
			"Test flag",
			"Usage of ToFloat64:\n  -contextTest float\n    \t[context] Test flag {TO_FLOAT64_CONTEXT_TEST} (default 12.34)\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToFloat64", flag.ContinueOnError)
			fg := NewFlag(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToFloat64(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToFloat64() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	var cases = []struct {
		intention    string
		prefix       string
		docPrefix    string
		name         string
		defaultValue bool
		label        string
		want         string
	}{
		{
			"simple",
			"",
			"cli",
			"test",
			false,
			"Test flag",
			"Usage of ToBool:\n  -test\n    \t[cli] Test flag {TO_BOOL_TEST}\n",
		},
		{
			"with prefix",
			"context",
			"cli",
			"test",
			true,
			"Test flag",
			"Usage of ToBool:\n  -contextTest\n    \t[context] Test flag {TO_BOOL_CONTEXT_TEST} (default true)\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToBool", flag.ContinueOnError)
			fg := NewFlag(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToBool(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToBool() = %s, want %s", result, testCase.want)
			}
		})
	}
}

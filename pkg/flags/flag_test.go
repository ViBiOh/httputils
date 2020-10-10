package flags

import (
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	type args struct {
		name      string
		value     interface{}
		overrides []Override
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
	}{
		{
			"empty",
			args{
				name:  "value",
				value: "empty",
			},
			"empty",
		},
		{
			"match",
			args{
				name:  "value",
				value: "empty",
				overrides: []Override{
					NewOverride("VALUE", "found"),
				},
			},
			"found",
		},
		{
			"no match",
			args{
				name:  "value",
				value: "empty",
				overrides: []Override{
					NewOverride("value1", "found"),
					NewOverride("value2", "found"),
				},
			},
			"empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := Default(tc.args.name, tc.args.value, tc.args.overrides); got != tc.want {
				t.Errorf("Default() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		docPrefix string
		want      *Flag
	}{
		{
			"simple",
			"new",
			"test",
			&Flag{
				prefix:    "New",
				docPrefix: "new",
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
			if result := New(testCase.prefix, testCase.docPrefix); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("New() = %#v, want %#v", result, testCase.want)
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
		{
			"env",
			"",
			"cli",
			"value",
			"default",
			"Test flag",
			"Usage of ToString:\n  -value string\n    \t[cli] Test flag {TO_STRING_VALUE} (default \"test\")\n",
		},
	}

	os.Setenv("TO_STRING_VALUE", "test")

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToString", flag.ContinueOnError)
			fg := New(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToString(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToString() = `%s`, want `%s`", result, testCase.want)
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
		{
			"env",
			"",
			"cli",
			"value",
			8000,
			"Test flag",
			"Usage of ToInt:\n  -value int\n    \t[cli] Test flag {TO_INT_VALUE} (default 6000)\n",
		},
		{
			"invalid env",
			"",
			"cli",
			"invalidValue",
			8000,
			"Test flag",
			"Usage of ToInt:\n  -invalidValue int\n    \t[cli] Test flag {TO_INT_INVALID_VALUE} (default 8000)\n",
		},
	}

	os.Setenv("TO_INT_VALUE", "6000")
	os.Setenv("TO_INT_INVALID_VALUE", "test")

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToInt", flag.ContinueOnError)
			fg := New(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToInt(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToInt() = `%s`, want `%s`", result, testCase.want)
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
		defaultValue interface{}
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
			"uint",
			"",
			"cli",
			"test",
			uint(10),
			"Test flag",
			"Usage of ToUint:\n  -test uint\n    \t[cli] Test flag {TO_UINT_TEST} (default 10)\n",
		},
		{
			"uint",
			"",
			"cli",
			"test",
			"test",
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
		{
			"env",
			"",
			"cli",
			"value",
			8000,
			"Test flag",
			"Usage of ToUint:\n  -value uint\n    \t[cli] Test flag {TO_UINT_VALUE} (default 6000)\n",
		},
		{
			"invalid env",
			"",
			"cli",
			"invalidValue",
			8000,
			"Test flag",
			"Usage of ToUint:\n  -invalidValue uint\n    \t[cli] Test flag {TO_UINT_INVALID_VALUE} (default 8000)\n",
		},
	}

	os.Setenv("TO_UINT_VALUE", "6000")
	os.Setenv("TO_UINT_INVALID_VALUE", "-6000")

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToUint", flag.ContinueOnError)
			fg := New(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToUint(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToUint() = `%s`, want `%s`", result, testCase.want)
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
		{
			"env",
			"",
			"cli",
			"value",
			12.34,
			"Test flag",
			"Usage of ToFloat64:\n  -value float\n    \t[cli] Test flag {TO_FLOAT64_VALUE} (default 34.56)\n",
		},
		{
			"invalid env",
			"",
			"cli",
			"invalidValue",
			12.34,
			"Test flag",
			"Usage of ToFloat64:\n  -invalidValue float\n    \t[cli] Test flag {TO_FLOAT64_INVALID_VALUE} (default 12.34)\n",
		},
	}

	os.Setenv("TO_FLOAT64_VALUE", "34.56")
	os.Setenv("TO_FLOAT64_INVALID_VALUE", "12.34.56")

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToFloat64", flag.ContinueOnError)
			fg := New(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToFloat64(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToFloat64() = `%s`, want `%s`", result, testCase.want)
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
		{
			"env",
			"",
			"cli",
			"value",
			true,
			"Test flag",
			"Usage of ToBool:\n  -value\n    \t[cli] Test flag {TO_BOOL_VALUE}\n",
		},
		{
			"invalid env",
			"",
			"cli",
			"invalidValue",
			true,
			"Test flag",
			"Usage of ToBool:\n  -invalidValue\n    \t[cli] Test flag {TO_BOOL_INVALID_VALUE} (default true)\n",
		},
	}

	os.Setenv("TO_BOOL_VALUE", "false")
	os.Setenv("TO_BOOL_INVALID_VALUE", "test")

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("ToBool", flag.ContinueOnError)
			fg := New(testCase.prefix, testCase.docPrefix).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
			fg.ToBool(fs)

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if result := writer.String(); result != testCase.want {
				t.Errorf("ToBool() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

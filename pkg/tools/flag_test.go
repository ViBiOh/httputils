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
		name         string
		defaultValue string
		label        string
		want         string
	}{
		{
			"simple",
			"test",
			"default",
			"Test flag",
			"Usage of toString:\n  -simpletest string\n    \t[simple] Test flag {TO_STRING_SIMPLETEST} (default \"default\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet("toString", flag.ContinueOnError)
			fg := NewFlag(testCase.intention, testCase.intention).Name(testCase.name).Default(testCase.defaultValue).Label(testCase.label)
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

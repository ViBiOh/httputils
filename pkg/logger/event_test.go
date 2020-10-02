package logger

import (
	"errors"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	type args struct {
		line string
	}

	var cases = []struct {
		intention string
		args      args
		want      level
		wantErr   error
	}{
		{
			"default value",
			args{
				line: "",
			},
			levelInfo,
			errors.New("invalid value ``"),
		},
		{
			"lower case value",
			args{
				line: "debug",
			},
			levelDebug,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := parseLevel(tc.args.line)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("parseLevel() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"nothing",
			args{
				content: "test",
			},
			"test",
		},
		{
			"complex",
			args{
				content: "Text with special character \"'\b\f\t\r\n.",
			},
			`Text with special character \"'\b\f\t\r\n.`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := EscapeString(tc.args.content); got != tc.want {
				t.Errorf("EscapeString() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkEscapeStringSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EscapeString("Text with simple character.")
	}
}

func BenchmarkEscapeStringComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EscapeString("Text with special character /\"'\b\f\t\r\n.")
	}
}

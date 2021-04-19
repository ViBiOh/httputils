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

package logger

import (
	"errors"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()

	type args struct {
		line string
	}

	cases := map[string]struct {
		args    args
		want    level
		wantErr error
	}{
		"default value": {
			args{
				line: "",
			},
			levelInfo,
			errors.New("invalid value ``"),
		},
		"lower case value": {
			args{
				line: "debug",
			},
			levelDebug,
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := parseLevel(testCase.args.line)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()),
				got != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("parseLevel() = (%d, `%s`), want (%d, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

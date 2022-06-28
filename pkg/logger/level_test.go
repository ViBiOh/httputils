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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			got, gotErr := parseLevel(tc.args.line)

			failed := false

			switch {
			case
				tc.wantErr == nil && gotErr != nil,
				tc.wantErr != nil && gotErr == nil,
				tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()),
				got != tc.want:
				failed = true
			}

			if failed {
				t.Errorf("parseLevel() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

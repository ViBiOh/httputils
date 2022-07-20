package bcrypt

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestFindBestCost(t *testing.T) {
	type args struct {
		maxDuration time.Duration
	}

	cases := map[string]struct {
		args    args
		want    int
		wantErr error
	}{
		"min cost": {
			args{
				maxDuration: 0,
			},
			bcrypt.MinCost,
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			got, gotErr := FindBestCost(tc.args.maxDuration)

			failed := false

			switch {
			case
				tc.wantErr == nil && gotErr != nil,
				tc.wantErr != nil && gotErr == nil,
				tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()),
				got != tc.want:
				failed = true
			}

			if failed {
				t.Errorf("FindBestCost() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

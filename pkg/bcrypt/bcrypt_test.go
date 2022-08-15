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

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := FindBestCost(testCase.args.maxDuration)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()),
				got != testCase.want:
				failed = true
			}

			if failed {
				t.Errorf("FindBestCost() = (%d, `%s`), want (%d, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

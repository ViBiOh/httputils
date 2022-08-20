package uuid

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want    string
		wantErr error
	}{
		"simple": {
			"f6a89f66-bece-4c93-85e5-4da6831b28fb",
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := New()

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()),
				len(got) != len(testCase.want):
				failed = true
			}

			if failed {
				t.Errorf("New() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

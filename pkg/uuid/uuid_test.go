package uuid

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	cases := map[string]struct {
		want    string
		wantErr error
	}{
		"simple": {
			"f6a89f66-bece-4c93-85e5-4da6831b28fb",
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			got, gotErr := New()

			failed := false

			switch {
			case
				tc.wantErr == nil && gotErr != nil,
				tc.wantErr != nil && gotErr == nil,
				tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()),
				len(got) != len(tc.want):
				failed = true
			}

			if failed {
				t.Errorf("New() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

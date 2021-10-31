package uuid

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	cases := []struct {
		intention string
		want      string
		wantErr   error
	}{
		{
			"simple",
			"f6a89f66-bece-4c93-85e5-4da6831b28fb",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New()

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if len(got) != len(tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("New() = (`%s`, `%s`), want (`%s`, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

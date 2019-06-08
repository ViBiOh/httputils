package uuid

import (
	"testing"
)

func TestNew(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
		wantErr   error
	}{
		{
			"should work",
			16*2 + 4,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := New()

			failed := false

			if err == nil && testCase.wantErr != nil {
				failed = true
			} else if err != nil && testCase.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != testCase.wantErr.Error() {
				failed = true
			} else if len(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("New() = (%#v, %#v), want (%#v, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

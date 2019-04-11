package uuid

import (
	"testing"
)

func Test_New(t *testing.T) {
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

	var failed bool

	for _, testCase := range cases {
		result, err := New()

		failed = false

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
			t.Errorf("%s\nNew() = (%+v, %+v), want (%+v, %+v)", testCase.intention, result, err, testCase.want, testCase.wantErr)
		}
	}
}

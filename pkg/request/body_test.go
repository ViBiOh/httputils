package request

import (
	"bytes"
	"net/http"
	"testing"
)

func Test_ReadBodyRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, `http://localhost`, bytes.NewBuffer([]byte(`Content from buffer`)))
	if err != nil {
		t.Errorf(`Unable to create test request: %v`, err)
	}

	var cases = []struct {
		request *http.Request
		want    string
		wantErr error
	}{
		{
			req,
			`Content from buffer`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := ReadBodyRequest(testCase.request)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if string(result) != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf(`ReadBodyRequest(%v) = (%v, %v), want (%v, %v)`, testCase.request, result, err, testCase.want, testCase.wantErr)
		}
	}
}
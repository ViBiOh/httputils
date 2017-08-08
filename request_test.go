package httputils

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestGetBasicAuth(t *testing.T) {
	var tests = []struct {
		username string
		password string
		want     string
	}{
		{
			``,
			``,
			`Basic Og==`,
		},
		{
			`admin`,
			`password`,
			`Basic YWRtaW46cGFzc3dvcmQ=`,
		},
	}

	for _, test := range tests {
		if result := GetBasicAuth(test.username, test.password); result != test.want {
			t.Errorf(`GetBasicAuth(%v, %v) = %v, want %v`, test.username, test.password, result, test.want)
		}
	}
}

func TestReadBody(t *testing.T) {
	var tests = []struct {
		body    io.ReadCloser
		want    string
		wantErr error
	}{
		{
			ioutil.NopCloser(bytes.NewBuffer([]byte(`Content from buffer`))),
			`Content from buffer`,
			nil,
		},
	}

	var failed bool

	for _, test := range tests {
		result, err := ReadBody(test.body)

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		} else if string(result) != test.want {
			failed = true
		}

		if failed {
			t.Errorf(`GetBasicAuth(%v) = (%v, %v), want %v`, test.body, result, err, test.want, test.wantErr)
		}
	}
}

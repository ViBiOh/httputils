package httputils

import (
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

package request

import (
	"testing"
)

func TestGenerateBasicAuth(t *testing.T) {
	var cases = []struct {
		username string
		password string
		want     string
	}{
		{
			"",
			"",
			"Basic Og==",
		},
		{
			"admin",
			"password",
			"Basic YWRtaW46cGFzc3dvcmQ=",
		},
	}

	for _, testCase := range cases {
		if result := GenerateBasicAuth(testCase.username, testCase.password); result != testCase.want {
			t.Errorf("GenerateBasicAuth(%#v, %#v) = %#v, want %#v", testCase.username, testCase.password, result, testCase.want)
		}
	}
}

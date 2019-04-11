package tools

import (
	"testing"
)

func Test_Sha1(t *testing.T) {
	var cases = []struct {
		intention string
		input     interface{}
		want      string
	}{
		{
			"should work with nil",
			nil,
			"3a9bcf8af38962fe340baa717773bf285f95121a",
		},
		{
			"should work with string",
			"Hello world",
			"7b502c3a1f48c8609ae212cdfb639dee39673f5e",
		},
	}

	for _, testCase := range cases {
		if result := Sha1(testCase.input); result != testCase.want {
			t.Errorf("%s\nSha1(%+v) = %+v, want %+v", testCase.intention, testCase.input, result, testCase.want)
		}
	}
}

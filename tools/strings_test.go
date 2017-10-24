package tools

import "testing"

func Test_ToCamel(t *testing.T) {
	var cases = []struct {
		intention string
		input     string
		want      string
	}{
		{
			`should work with empty string`,
			``,
			``,
		},
		{
			`should work with lower case string`,
			`test`,
			`test`,
		},
		{
			`should work with regular string`,
			`OhPleaseFormatMe`,
			`ohPleaseFormatMe`,
		},
	}

	for _, testCase := range cases {
		if result := ToCamel(testCase.input); result != testCase.want {
			t.Errorf("%v\nToCamel(%v) = %v, want %v", testCase.intention, testCase.input, result, testCase.want)
		}
	}
}

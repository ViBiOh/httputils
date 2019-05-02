package tools

import "testing"

func TestToCamel(t *testing.T) {
	var cases = []struct {
		intention string
		input     string
		want      string
	}{
		{
			"should work with empty string",
			"",
			"",
		},
		{
			"should work with lower case string",
			"test",
			"test",
		},
		{
			"should work with regular string",
			"OhPleaseFormatMe",
			"ohPleaseFormatMe",
		},
	}

	for _, testCase := range cases {
		if result := ToCamel(testCase.input); result != testCase.want {
			t.Errorf("%v\nToCamel(%v) = %v, want %v", testCase.intention, testCase.input, result, testCase.want)
		}
	}
}

func TestIncludesString(t *testing.T) {
	var cases = []struct {
		intention string
		array     []string
		lookup    string
		want      bool
	}{
		{
			"should work with nil params",
			nil,
			"",
			false,
		},
		{
			"should work with found value",
			[]string{"hello", "world"},
			"world",
			true,
		},
		{
			"should work with not found value",
			[]string{"hello", "world"},
			"bob",
			false,
		},
	}

	for _, testCase := range cases {
		if result := IncludesString(testCase.array, testCase.lookup); result != testCase.want {
			t.Errorf("%s\nIncludes(%+v, `%s`) = %+v, want %+v", testCase.intention, testCase.array, testCase.lookup, result, testCase.want)
		}
	}
}

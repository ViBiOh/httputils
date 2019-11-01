package flags

import "testing"

func TestFirstLowerCase(t *testing.T) {
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
		t.Run(testCase.intention, func(t *testing.T) {
			if result := FirstLowerCase(testCase.input); result != testCase.want {
				t.Errorf("FirstUpperCase(%#v) = %#v, want %#v", testCase.input, result, testCase.want)
			}
		})
	}
}

func TestFirstUpperCase(t *testing.T) {
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
			"Test",
		},
		{
			"should work with regular string",
			"OhPleaseFormatMe",
			"OhPleaseFormatMe",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := FirstUpperCase(testCase.input); result != testCase.want {
				t.Errorf("FirstUpperCase(%#v) = %#v, want %#v", testCase.input, result, testCase.want)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
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
			"should work with basic string",
			"test",
			"test",
		},
		{
			"should work with upper case starting string",
			"OhPleaseFormatMe",
			"Oh_Please_Format_Me",
		},
		{
			"should work with camelCase string",
			"listCount",
			"list_Count",
		},
		{
			"should work with dash bestween word",
			"List-Of_thing",
			"List_Of_thing",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := SnakeCase(testCase.input); result != testCase.want {
				t.Errorf("SnakeCase(%#v) = %#v, want %#v", testCase.input, result, testCase.want)
			}
		})
	}
}

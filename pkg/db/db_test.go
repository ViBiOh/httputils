package db

import (
	"fmt"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string host param to flags`,
			`host`,
			`*string`,
		},
		{
			`should add string port param to flags`,
			`port`,
			`*string`,
		},
		{
			`should add string user param to flags`,
			`user`,
			`*string`,
		},
		{
			`should add string pass param to flags`,
			`pass`,
			`*string`,
		},
		{
			`should add string name param to flags`,
			`name`,
			`*string`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
		}
	}
}

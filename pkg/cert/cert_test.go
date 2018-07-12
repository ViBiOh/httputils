package cert

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
			`should add string cert param to flags`,
			`cert`,
			`*string`,
		},
		{
			`should add string key param to flags`,
			`key`,
			`*string`,
		},
		{
			`should add string organization param to flags`,
			`organization`,
			`*string`,
		},
		{
			`should add string hosts param to flags`,
			`hosts`,
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

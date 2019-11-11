package breaksync

import "testing"

func TestComputeSynchro(t *testing.T) {
	simple := NewSource(nil, nil, nil)
	simple.currentKey = "AAAAA00000"

	var cases = []struct {
		intention string
		instance  *Source
		input     string
		want      bool
	}{
		{
			"simple",
			simple,
			"AAAAA00000",
			true,
		},
		{
			"substring",
			simple,
			"AAAAA",
			true,
		},
		{
			"extrastring",
			simple,
			"AAAAA00000zzzzz",
			true,
		},
		{
			"unmatch",
			simple,
			"AAAAA00001",
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			testCase.instance.computeSynchro(testCase.input)
			if testCase.instance.synchronized != testCase.want {
				t.Errorf("computeSynchro() = %t, want %t", testCase.instance.synchronized, testCase.want)
			}
		})
	}
}

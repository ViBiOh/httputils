package breaksync

import "testing"

func TestCompute(t *testing.T) {
	var cases = []struct {
		intention string
		instance  *Rupture
		current   string
		next      string
		force     bool
		want      bool
	}{
		{
			"simple",
			NewRupture("simple", RuptureExtractSimple),
			"A0",
			"A1",
			false,
			true,
		},
		{
			"equal",
			NewRupture("simple", RuptureExtractSimple),
			"A0",
			"A0",
			false,
			false,
		},
		{
			"forced",
			NewRupture("simple", RuptureExtractSimple),
			"A0",
			"A0",
			true,
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.instance.compute(testCase.current, testCase.next, testCase.force); result != testCase.want {
				t.Errorf("Compute() = %t, want %t", result, testCase.want)
			}
		})
	}
}

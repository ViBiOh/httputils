package breaksync

import "testing"

func TestCompute(t *testing.T) {
	cases := map[string]struct {
		instance *Rupture
		current  string
		next     string
		force    bool
		want     bool
	}{
		"simple": {
			NewRupture("simple", Identity),
			"A0",
			"A1",
			false,
			true,
		},
		"equal": {
			NewRupture("simple", Identity),
			"A0",
			"A0",
			false,
			false,
		},
		"forced": {
			NewRupture("simple", Identity),
			"A0",
			"A0",
			true,
			true,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := testCase.instance.compute(testCase.current, testCase.next, testCase.force); result != testCase.want {
				t.Errorf("Compute() = %t, want %t", result, testCase.want)
			}
		})
	}
}

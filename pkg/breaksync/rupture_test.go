package breaksync

import "testing"

func TestCompute(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance *Rupture
		current  []byte
		next     []byte
		force    bool
		want     bool
	}{
		"simple": {
			NewRupture("simple", RuptureIdentity),
			[]byte("A0"),
			[]byte("A1"),
			false,
			true,
		},
		"equal": {
			NewRupture("simple", RuptureIdentity),
			[]byte("A0"),
			[]byte("A0"),
			false,
			false,
		},
		"forced": {
			NewRupture("simple", RuptureIdentity),
			[]byte("A0"),
			[]byte("A0"),
			true,
			true,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := testCase.instance.compute(testCase.current, testCase.next, testCase.force); result != testCase.want {
				t.Errorf("Compute() = %t, want %t", result, testCase.want)
			}
		})
	}
}

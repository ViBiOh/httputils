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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if result := tc.instance.compute(tc.current, tc.next, tc.force); result != tc.want {
				t.Errorf("Compute() = %t, want %t", result, tc.want)
			}
		})
	}
}

package sha

import "testing"

func TestNew(t *testing.T) {
	type args struct {
		o interface{}
	}

	value := "test"

	cases := []struct {
		intention string
		args      args
		want      string
	}{
		{
			"simple",
			args{
				o: value,
			},
			"5006d6f8302000e8b87fef5c50c071d6d97b4e88",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := New(tc.args.o); got != tc.want {
				t.Errorf("New() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(New)
	}
}

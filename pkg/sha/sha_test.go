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

func TestStream(t *testing.T) {
	type args struct {
		o []interface{}
	}

	value := "test"

	cases := []struct {
		intention string
		args      args
		want      string
	}{
		{
			"empty",
			args{
				o: nil,
			},
			"da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			"simple",
			args{
				o: []interface{}{value},
			},
			"5006d6f8302000e8b87fef5c50c071d6d97b4e88",
		},
		{
			"multiple",
			args{
				o: []interface{}{value, value},
			},
			"6d5182a16753e136eae9f72329c84e3dc9e0ae67",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			stream := Stream()

			for _, item := range tc.args.o {
				stream.Write(item)
			}

			if got := stream.Sum(); got != tc.want {
				t.Errorf("Stream() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(New)
	}
}

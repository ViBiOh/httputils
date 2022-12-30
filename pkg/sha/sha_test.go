package sha

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	type args struct {
		o any
	}

	value := "test"

	cases := map[string]struct {
		args args
		want string
	}{
		"simple": {
			args{
				o: value,
			},
			"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := New(testCase.args.o); got != testCase.want {
				t.Errorf("New() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestStream(t *testing.T) {
	t.Parallel()

	type args struct {
		o []any
	}

	value := "test"

	cases := map[string]struct {
		args args
		want string
	}{
		"empty": {
			args{
				o: nil,
			},
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		"simple": {
			args{
				o: []any{value},
			},
			"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		"multiple": {
			args{
				o: []any{value, value},
			},
			"37268335dd6931045bdcdf92623ff819a64244b53d0e746d438797349d4da578",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			stream := Stream()

			for _, item := range testCase.args.o {
				stream.Write(item)
			}

			if got := stream.Sum(); got != testCase.want {
				t.Errorf("Stream() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	type testStruct struct {
		ID int
	}

	item := testStruct{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(item)
	}
}

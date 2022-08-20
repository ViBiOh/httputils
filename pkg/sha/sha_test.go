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
			"4d967a30111bf29f0eba01c448b375c1629b2fed01cdfcc3aed91f1b57d5dd5e",
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
			"4d967a30111bf29f0eba01c448b375c1629b2fed01cdfcc3aed91f1b57d5dd5e",
		},
		"multiple": {
			args{
				o: []any{value, value},
			},
			"bfc919009a7b3b1f6b7ed14e39c8b0194782f1caf3067ce874953cc887f1ac4f",
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

	for i := 0; i < b.N; i++ {
		New(item)
	}
}

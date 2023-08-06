package hash

import "testing"

func TestString(t *testing.T) {
	t.Parallel()

	type args struct {
		o string
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
			"9ec9f7918d7dfc40",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := String(testCase.args.o); got != testCase.want {
				t.Errorf("String() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestHash(t *testing.T) {
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
			"9ec9f7918d7dfc40",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := Hash(testCase.args.o); got != testCase.want {
				t.Errorf("Hash() = `%s`, want `%s`", got, testCase.want)
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
			"2d06800538d394c2",
		},
		"simple": {
			args{
				o: []any{value},
			},
			"9ec9f7918d7dfc40",
		},
		"multiple": {
			args{
				o: []any{value, value},
			},
			"13d5f5d1923ebbf0",
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
		Hash(item)
	}
}

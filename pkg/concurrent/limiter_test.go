package concurrent

import (
	"testing"
)

func TestLimitedGo(t *testing.T) {
	t.Parallel()

	type args struct {
		funcs []func()
	}

	cases := map[string]struct {
		instance *Limited
		args     args
	}{
		"simple": {
			NewLimited(2),
			args{
				funcs: []func(){
					func() {},
					func() {},
				},
			},
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			for _, f := range testCase.args.funcs {
				testCase.instance.Go(f)
			}

			testCase.instance.Wait()
		})
	}
}

func BenchmarkLimited(b *testing.B) {
	funcs := []func(){
		func() {},
		func() {},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instance := NewLimited(2)

		for _, f := range funcs {
			instance.Go(f)
		}

		instance.Wait()
	}
}

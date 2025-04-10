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
		instance *Limiter
		args     args
	}{
		"simple": {
			NewLimiter(-1),
			args{
				funcs: []func(){
					func() {},
					func() {},
				},
			},
		},
		"two": {
			NewLimiter(2),
			args{
				funcs: []func(){
					func() {},
					func() {},
				},
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			for _, f := range testCase.args.funcs {
				testCase.instance.Go(f)
			}

			testCase.instance.Wait()
		})
	}
}

func BenchmarkLimiter(b *testing.B) {
	funcs := []func(){
		func() {},
		func() {},
	}

	for b.Loop() {
		instance := NewLimiter(2)

		for _, f := range funcs {
			instance.Go(f)
		}

		instance.Wait()
	}
}

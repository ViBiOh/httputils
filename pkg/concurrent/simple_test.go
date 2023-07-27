package concurrent

import (
	"testing"
)

func TestSimpleGo(t *testing.T) {
	t.Parallel()

	type args struct {
		funcs []func()
	}

	cases := map[string]struct {
		instance Runner
		args     args
	}{
		"simple": {
			NewSimple(),
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

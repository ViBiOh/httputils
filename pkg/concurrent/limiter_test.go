package concurrent

import (
	"testing"
)

func TestLimitedGo(t *testing.T) {
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
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			for _, f := range testCase.args.funcs {
				testCase.instance.Go(f)
			}

			testCase.instance.Wait()
		})
	}
}

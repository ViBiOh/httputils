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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			for _, f := range tc.args.funcs {
				tc.instance.Go(f)
			}

			tc.instance.Wait()
		})
	}
}

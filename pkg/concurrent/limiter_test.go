package concurrent

import (
	"testing"
)

func TestLimitedGo(t *testing.T) {
	type args struct {
		funcs []func()
	}

	cases := []struct {
		intention string
		instance  *Limited
		args      args
	}{
		{
			"simple",
			NewLimited(2),
			args{
				funcs: []func(){
					func() {},
					func() {},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			for _, f := range tc.args.funcs {
				tc.instance.Go(f)
			}

			tc.instance.Wait()
		})
	}
}

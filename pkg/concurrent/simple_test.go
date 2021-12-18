package concurrent

import (
	"testing"
)

func TestSimpleGo(t *testing.T) {
	type args struct {
		funcs []func()
	}

	cases := []struct {
		intention string
		instance  *Simple
		args      args
	}{
		{
			"simple",
			NewSimple(),
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

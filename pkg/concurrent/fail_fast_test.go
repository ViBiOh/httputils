package concurrent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestFailFastGo(t *testing.T) {
	type args struct {
		funcs []func() error
	}

	cases := []struct {
		intention string
		instance  *FailFast
		args      args
		wantErr   error
	}{
		{
			"no error",
			NewFailFast(2),
			args{
				funcs: []func() error{
					func() error { return nil },
					func() error { return nil },
				},
			},
			nil,
		},
		{
			"simple",
			NewFailFast(2),
			args{
				funcs: []func() error{
					func() error { return nil },
					func() error { return errors.New("failed one") },
				},
			},
			errors.New("failed one"),
		},
		{
			"two errors",
			NewFailFast(1),
			args{
				funcs: []func() error{
					func() error { return errors.New("failed one") },
					func() error { return errors.New("failed two") },
				},
			},
			errors.New("failed one"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if tc.intention != "no error" {
				tc.instance.WithContext(context.Background())
			}

			for _, f := range tc.args.funcs {
				tc.instance.Go(f)
			}

			gotErr := tc.instance.Wait()

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Go() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

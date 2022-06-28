package concurrent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestWithContext(t *testing.T) {
	type args struct {
		contexts []context.Context
	}

	cases := map[string]struct {
		instance *FailFast
		args     args
		wantErr  error
	}{
		"simple": {
			NewFailFast(1),
			args{
				contexts: []context.Context{
					context.Background(),
				},
			},
			nil,
		},
		"double": {
			NewFailFast(1),
			args{
				contexts: []context.Context{
					context.Background(),
					context.TODO(),
				},
			},
			errors.New("panic"),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			var gotErr error

			func() {
				defer func() {
					if e := recover(); e != nil {
						gotErr = errors.New("panic")
					}
				}()

				for _, ctx := range tc.args.contexts {
					tc.instance.WithContext(ctx)
				}
			}()

			failed := false

			switch {
			case
				tc.wantErr == nil && gotErr != nil,
				tc.wantErr != nil && gotErr == nil,
				tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("WithContext() = %s, want %s", gotErr, tc.wantErr)
			}
		})
	}
}

func TestFailFastGo(t *testing.T) {
	type args struct {
		funcs []func() error
	}

	cases := map[string]struct {
		instance *FailFast
		args     args
		wantErr  error
	}{
		"no error": {
			NewFailFast(2),
			args{
				funcs: []func() error{
					func() error { return nil },
					func() error { return nil },
				},
			},
			nil,
		},
		"simple": {
			NewFailFast(2),
			args{
				funcs: []func() error{
					func() error { return nil },
					func() error { return errors.New("failed one") },
				},
			},
			errors.New("failed one"),
		},
		"two errors": {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if intention != "no error" {
				tc.instance.WithContext(context.Background())
			}

			for _, f := range tc.args.funcs {
				tc.instance.Go(f)
			}

			gotErr := tc.instance.Wait()

			failed := false

			switch {
			case
				tc.wantErr == nil && gotErr != nil,
				tc.wantErr != nil && gotErr == nil,
				tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Go() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

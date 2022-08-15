package concurrent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestWithContext(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			var gotErr error

			func() {
				defer func() {
					if e := recover(); e != nil {
						gotErr = errors.New("panic")
					}
				}()

				for _, ctx := range testCase.args.contexts {
					testCase.instance.WithContext(ctx)
				}
			}()

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("WithContext() = %s, want %s", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestFailFastGo(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if intention != "no error" {
				testCase.instance.WithContext(context.Background())
			}

			for _, f := range testCase.args.funcs {
				testCase.instance.Go(f)
			}

			gotErr := testCase.instance.Wait()

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Go() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

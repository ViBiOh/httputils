package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestExclusive(t *testing.T) {
	t.Parallel()

	lockName := "lock"
	timeout := time.Minute

	cases := map[string]struct {
		timeout      time.Duration
		cancelAction bool
		setNX        *redis.BoolCmd
		release      *redis.Cmd
		action       func(context.Context) error
		wantAcquired bool
		wantErrs     []error
	}{
		"acquired and released": {
			timeout:      timeout,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(int64(1), nil),
			action:       func(context.Context) error { return nil },
			wantAcquired: true,
		},
		"already held": {
			timeout: timeout,
			setNX:   redis.NewBoolResult(false, nil),
		},
		"acquire error": {
			timeout:  timeout,
			setNX:    redis.NewBoolResult(false, errors.New("down")),
			wantErrs: []error{errors.New("exec setnx: down")},
		},
		"action error is returned and the lock released": {
			timeout:      timeout,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(int64(1), nil),
			action:       func(context.Context) error { return errors.New("failed") },
			wantAcquired: true,
			wantErrs:     []error{errors.New("failed")},
		},
		"lock expired before release": {
			timeout:      timeout,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(int64(0), nil),
			action:       func(context.Context) error { return nil },
			wantAcquired: true,
			wantErrs:     []error{ErrLockExpired},
		},
		"release error": {
			timeout:      timeout,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(nil, errors.New("down")),
			action:       func(context.Context) error { return nil },
			wantAcquired: true,
			wantErrs:     []error{errors.New("exec release: down")},
		},
		"action and release errors are both returned": {
			timeout:      timeout,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(int64(0), nil),
			action:       func(context.Context) error { return errors.New("failed") },
			wantAcquired: true,
			wantErrs:     []error{errors.New("failed"), ErrLockExpired},
		},
		"released after the caller is cancelled": {
			timeout:      timeout,
			cancelAction: true,
			setNX:        redis.NewBoolResult(true, nil),
			release:      redis.NewCmdResult(int64(1), nil),
			action:       func(ctx context.Context) error { return ctx.Err() },
			wantAcquired: true,
			wantErrs:     []error{context.Canceled},
		},
		"zero timeout": {
			timeout:  0,
			wantErrs: []error{ErrInvalidTimeout},
		},
		"negative timeout": {
			timeout:  -time.Second,
			wantErrs: []error{ErrInvalidTimeout},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			client := mocks.NewLockClient(ctrl)

			var token string

			if testCase.setNX != nil {
				client.EXPECT().SetNX(gomock.Any(), lockName, gomock.Any(), testCase.timeout).DoAndReturn(
					func(_ context.Context, _ string, value any, _ time.Duration) *redis.BoolCmd {
						token, _ = value.(string)

						return testCase.setNX
					},
				)
			}

			if testCase.release != nil {
				client.EXPECT().Eval(gomock.Any(), releaseScript, []string{lockName}, gomock.Any()).DoAndReturn(
					func(ctx context.Context, _ string, _ []string, args ...any) *redis.Cmd {
						assert.NoError(t, ctx.Err(), "release must not use a cancelled context")
						assert.Equal(t, []any{token}, args, "release must use the token set by the lock")

						return testCase.release
					},
				)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			action := func(actionCtx context.Context) error {
				if testCase.cancelAction {
					cancel()
				}

				return testCase.action(actionCtx)
			}

			if testCase.action == nil {
				action = func(context.Context) error {
					t.Error("action must not run")

					return nil
				}
			}

			acquired, err := exclusive(ctx, client, lockName, testCase.timeout, action)

			assert.Equal(t, testCase.wantAcquired, acquired)

			if len(testCase.wantErrs) == 0 {
				assert.NoError(t, err)

				return
			}

			for _, wantErr := range testCase.wantErrs {
				if errors.Is(wantErr, ErrLockExpired) || errors.Is(wantErr, ErrInvalidTimeout) || errors.Is(wantErr, context.Canceled) {
					assert.ErrorIs(t, err, wantErr)
				} else {
					assert.ErrorContains(t, err, wantErr.Error())
				}
			}
		})
	}
}

func TestExclusiveToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := mocks.NewLockClient(ctrl)

	tokens := map[string]struct{}{}

	client.EXPECT().SetNX(gomock.Any(), "lock", gomock.Any(), time.Minute).DoAndReturn(
		func(_ context.Context, _ string, value any, _ time.Duration) *redis.BoolCmd {
			token, _ := value.(string)
			tokens[token] = struct{}{}

			return redis.NewBoolResult(true, nil)
		},
	).Times(50)

	client.EXPECT().Eval(gomock.Any(), releaseScript, []string{"lock"}, gomock.Any()).Return(redis.NewCmdResult(int64(1), nil)).Times(50)

	for range 50 {
		_, err := exclusive(context.Background(), client, "lock", time.Minute, func(context.Context) error { return nil })

		assert.NoError(t, err)
	}

	assert.Len(t, tokens, 50)

	for token := range tokens {
		assert.GreaterOrEqual(t, len(token), 26)
	}
}

func TestExclusiveActionDeadline(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := mocks.NewLockClient(ctrl)

	client.EXPECT().SetNX(gomock.Any(), "lock", gomock.Any(), time.Minute).Return(redis.NewBoolResult(true, nil))
	client.EXPECT().Eval(gomock.Any(), releaseScript, []string{"lock"}, gomock.Any()).Return(redis.NewCmdResult(int64(1), nil))

	_, err := exclusive(context.Background(), client, "lock", time.Minute, func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()

		assert.True(t, ok)
		assert.False(t, deadline.After(time.Now().Add(time.Minute)), "the action must stop before the lock expires")

		return nil
	})

	assert.NoError(t, err)
}

package redis

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate go tool "go.uber.org/mock/mockgen" -source $GOFILE -destination ../mocks/$GOFILE -package mocks -mock_names lockClient=LockClient

const (
	releaseTimeout = time.Second * 5

	releaseScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
)

var (
	ErrInvalidTimeout = errors.New("exclusive lock requires a positive timeout")
	ErrLockExpired    = errors.New("exclusive lock expired before being released")
)

type lockClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd
}

func (s *Service) Exclusive(ctx context.Context, name string, timeout time.Duration, action func(context.Context) error) (bool, error) {
	return exclusive(ctx, s.client, name, timeout, action)
}

// The lock holds a random token so that a holder whose lock expired cannot release the lock of the next holder.
func exclusive(ctx context.Context, client lockClient, name string, timeout time.Duration, action func(context.Context) error) (acquired bool, err error) {
	if timeout <= 0 {
		return false, ErrInvalidTimeout
	}

	token := rand.Text()
	started := time.Now()

	acquired, err = client.SetNX(ctx, name, token, timeout).Result()
	if err != nil {
		return false, fmt.Errorf("exec setnx: %w", err)
	}

	if !acquired {
		return false, nil
	}

	defer func() {
		err = errors.Join(err, release(ctx, client, name, token))
	}()

	actionCtx, cancel := context.WithDeadline(ctx, started.Add(timeout))
	defer cancel()

	return true, action(actionCtx)
}

// The release must not depend on the caller's context, otherwise a cancelled caller keeps the lock until it expires.
func release(ctx context.Context, client lockClient, name, token string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), releaseTimeout)
	defer cancel()

	released, err := client.Eval(ctx, releaseScript, []string{name}, token).Int()
	if err != nil {
		return fmt.Errorf("exec release: %w", err)
	}

	if released == 0 {
		return ErrLockExpired
	}

	return nil
}

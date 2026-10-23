package cache

import (
	"context"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"go.uber.org/mock/gomock"
)

func TestExtend(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		build      func(client RedisClient) *TTLExtender
		wantExpire bool
	}{
		"ttl set before the extender": {
			func(client RedisClient) *TTLExtender {
				cache := New[string, string](client, func(id string) string { return id }, nil, nil)

				return cache.WithTTL(time.Hour).WithExtendOnHit(context.Background(), 0, 0).extender
			},
			true,
		},
		"ttl set after the extender": {
			func(client RedisClient) *TTLExtender {
				cache := New[string, string](client, func(id string) string { return id }, nil, nil)

				return cache.WithExtendOnHit(context.Background(), 0, 0).WithTTL(time.Hour).extender
			},
			true,
		},
		"no ttl": {
			func(client RedisClient) *TTLExtender {
				cache := New[string, string](client, func(id string) string { return id }, nil, nil)

				return cache.WithExtendOnHit(context.Background(), 0, 0).extender
			},
			false,
		},
		"negative ttl": {
			func(client RedisClient) *TTLExtender {
				return NewExtender(-time.Second, 0, 0, client)
			},
			false,
		},
		"batched, ttl set after the extender": {
			func(client RedisClient) *TTLExtender {
				cache := New[string, string](client, func(id string) string { return id }, nil, nil)

				return cache.WithExtendOnHit(context.Background(), time.Hour, 1).WithTTL(time.Hour).extender
			},
			true,
		},
		"batched, no ttl": {
			func(client RedisClient) *TTLExtender {
				cache := New[string, string](client, func(id string) string { return id }, nil, nil)

				return cache.WithExtendOnHit(context.Background(), time.Hour, 1).extender
			},
			false,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			client := mocks.NewRedisClient(ctrl)
			client.EXPECT().Enabled().Return(true).AnyTimes()

			if testCase.wantExpire {
				client.EXPECT().Expire(gomock.Any(), time.Hour, gomock.InAnyOrder([]string{"first", "second"})).Return(nil)
			}

			extender := testCase.build(client)

			if err := extender.Extend(context.Background(), "first", "second"); err != nil {
				t.Errorf("Extend() = `%s`, want no error", err)
			}
		})
	}
}

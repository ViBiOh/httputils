package cache_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEvictOnSuccess(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"error": {
			args{
				err: errors.New("update failed"),
			},
			errors.New("update failed"),
		},
		"evict": {
			args{
				err: nil,
			},
			nil,
		},
		"evict error": {
			args{
				err: nil,
			},
			errors.New("evict key `8000` from cache"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRedisClient := mocks.NewRedisClient(ctrl)
			mockRedisClient.EXPECT().Enabled().Return(true)

			switch intention {
			case "evict":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			case "evict error":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("redis failed"))
			}

			instance := cache.New(mockRedisClient, strconv.Itoa, func(_ context.Context, id int) (string, error) {
				return "hello", nil
			}, nil)

			gotErr := instance.EvictOnSuccess(context.Background(), 8000, testCase.args.err)
			if testCase.wantErr == nil {
				assert.NoError(t, gotErr)
			} else {
				assert.ErrorContains(t, gotErr, testCase.wantErr.Error())
			}
		})
	}
}

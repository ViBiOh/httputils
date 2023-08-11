package cache

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"go.uber.org/mock/gomock"
)

type cacheableItem struct {
	ID int `json:"id"`
}

type jsonErrorItem struct {
	ID    int           `json:"id"`
	Value func() string `json:"value"`
}

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		key      int
		onMiss   func(context.Context, int) (cacheableItem, error)
		duration time.Duration
	}

	cases := map[string]struct {
		args    args
		want    cacheableItem
		wantErr error
	}{
		"cache error": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (cacheableItem, error) {
					return cacheableItem{
						ID: 8000,
					}, nil
				},
				duration: time.Minute,
			},
			cacheableItem{
				ID: 8000,
			},
			nil,
		},
		"cache unmarshal": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (cacheableItem, error) {
					return cacheableItem{
						ID: 8000,
					}, nil
				},
				duration: time.Minute,
			},
			cacheableItem{
				ID: 8000,
			},
			nil,
		},
		"cached": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (cacheableItem, error) {
					return cacheableItem{}, nil
				},
			},
			cacheableItem{
				ID: 8000,
			},
			nil,
		},
		"store error": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (cacheableItem, error) {
					return cacheableItem{
						ID: 8000,
					}, nil
				},
				duration: time.Minute,
			},
			cacheableItem{
				ID: 8000,
			},
			nil,
		},
		"slow cache": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (cacheableItem, error) {
					return cacheableItem{
						ID: 9000,
					}, nil
				},
			},
			cacheableItem{
				ID: 9000,
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)
			mockRedisClient.EXPECT().Enabled().Return(true)

			switch intention {
			case "cache error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return(nil, errors.New("cache failed"))
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "cache unmarshal":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return([]byte("{"), nil)
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "cached":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return([]byte(`{"id":8000}`), nil)
				mockRedisClient.EXPECT().Expire(gomock.Any(), testCase.args.duration, gomock.Any()).Return(nil)
			case "store error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("store error"))
			case "slow cache":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key string) ([]byte, error) {
					select {
					case <-time.NewTimer(time.Second).C:
						return []byte(`{"id":8000}`), nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				})
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			instance := New(mockRedisClient, strconv.Itoa, testCase.args.onMiss, testCase.args.duration, -1, nil)

			got, gotErr := instance.Get(context.Background(), testCase.args.key)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()),
				!reflect.DeepEqual(got, testCase.want):
				failed = true
			}

			time.Sleep(time.Millisecond * 200)

			if failed {
				t.Errorf("Retrieve() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGetError(t *testing.T) {
	t.Parallel()

	type args struct {
		key      int
		onMiss   func(context.Context, int) (jsonErrorItem, error)
		duration time.Duration
	}

	funcValue := func() string {
		return "fail"
	}

	cases := map[string]struct {
		args    args
		want    jsonErrorItem
		wantErr error
	}{
		"marshal error": {
			args{
				key: 8000,
				onMiss: func(_ context.Context, id int) (jsonErrorItem, error) {
					return jsonErrorItem{
						ID:    id,
						Value: funcValue,
					}, nil
				},
				duration: time.Minute,
			},
			jsonErrorItem{
				ID:    8000,
				Value: funcValue,
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)
			mockRedisClient.EXPECT().Enabled().Return(true)

			switch intention {
			case "marshal error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return(nil, nil)
			}

			instance := New(mockRedisClient, strconv.Itoa, testCase.args.onMiss, testCase.args.duration, -1, nil)

			got, gotErr := instance.Get(context.Background(), testCase.args.key)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()),
				got.ID != testCase.want.ID:
				failed = true
			}

			time.Sleep(time.Millisecond * 200)

			if failed {
				t.Errorf("Retrieve() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

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
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch intention {
			case "evict":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			case "evict error":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("redis failed"))
			}

			instance := App[int, jsonErrorItem]{
				write: mockRedisClient,
				toKey: strconv.Itoa,
			}

			gotErr := instance.EvictOnSuccess(context.Background(), 8000, testCase.args.err)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("EvictOnSuccess() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

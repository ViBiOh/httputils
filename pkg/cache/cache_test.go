package cache

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/golang/mock/gomock"
)

type cacheableItem struct {
	ID int `json:"id"`
}

type jsonErrorItem struct {
	ID    int           `json:"id"`
	Value func() string `json:"value"`
}

func TestRetrieve(t *testing.T) {
	type args struct {
		key      string
		onMiss   func(_ context.Context) (cacheableItem, error)
		duration time.Duration
	}

	cases := map[string]struct {
		args    args
		want    cacheableItem
		wantErr error
	}{
		"cache error": {
			args{
				key: "8000",
				onMiss: func(_ context.Context) (cacheableItem, error) {
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
				key: "8000",
				onMiss: func(_ context.Context) (cacheableItem, error) {
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
				key: "8000",
				onMiss: func(_ context.Context) (cacheableItem, error) {
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
				key: "8000",
				onMiss: func(_ context.Context) (cacheableItem, error) {
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
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch intention {
			case "cache error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return("", errors.New("cache failed"))
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "cache unmarshal":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return("{", nil)
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "cached":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return(`{"id":8000}`, nil)
			case "store error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return("", nil)
				mockRedisClient.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("store error"))
			}

			got, gotErr := Retrieve(context.TODO(), mockRedisClient, tc.args.key, tc.args.onMiss, tc.args.duration)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			time.Sleep(time.Millisecond * 200)

			if failed {
				t.Errorf("Retrieve() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestRetrieveError(t *testing.T) {
	type args struct {
		key      string
		onMiss   func(_ context.Context) (jsonErrorItem, error)
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
				key: "8000",
				onMiss: func(_ context.Context) (jsonErrorItem, error) {
					return jsonErrorItem{
						ID:    8000,
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch intention {
			case "marshal error":
				mockRedisClient.EXPECT().Load(gomock.Any(), gomock.Any()).Return("", nil)
			}

			got, gotErr := Retrieve(context.TODO(), mockRedisClient, tc.args.key, tc.args.onMiss, tc.args.duration)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got.ID != tc.want.ID {
				failed = true
			}

			time.Sleep(time.Millisecond * 200)

			if failed {
				t.Errorf("Retrieve() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestEvictOnSuccess(t *testing.T) {
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
			errors.New("unable to evict key `key` from cache"),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch intention {
			case "evict":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			case "evict error":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("redis failed"))
			}

			gotErr := EvictOnSuccess(context.Background(), mockRedisClient, "key", tc.args.err)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("EvictOnSuccess() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

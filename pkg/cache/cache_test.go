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

func TestRetrieve(t *testing.T) {
	type args struct {
		key      string
		item     interface{}
		onMiss   func() (interface{}, error)
		duration time.Duration
	}

	cases := []struct {
		intention string
		args      args
		want      interface{}
		wantErr   error
	}{
		{
			"cache error",
			args{
				key: "8000",
				onMiss: func() (interface{}, error) {
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
		{
			"cache unmarshal",
			args{
				key: "8000",
				onMiss: func() (interface{}, error) {
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
		{
			"cached",
			args{
				key:  "8000",
				item: &cacheableItem{},
			},
			&cacheableItem{
				ID: 8000,
			},
			nil,
		},
		{
			"store error",
			args{
				key: "8000",
				onMiss: func() (interface{}, error) {
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch tc.intention {
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

			got, gotErr := Retrieve(context.TODO(), mockRedisClient, tc.args.key, tc.args.item, tc.args.onMiss, tc.args.duration)

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
				t.Errorf("Retrieve() = (%t, `%s`), want (%t, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestOnModify(t *testing.T) {
	type args struct {
		err error
	}

	cases := []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"error",
			args{
				err: errors.New("update failed"),
			},
			errors.New("update failed"),
		},
		{
			"evict",
			args{
				err: nil,
			},
			nil,
		},
		{
			"evict error",
			args{
				err: nil,
			},
			errors.New("unable to evict key `key` from cache"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedisClient := mocks.NewRedisClient(ctrl)

			switch tc.intention {
			case "evict":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			case "evict error":
				mockRedisClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("redis failed"))
			}

			gotErr := OnModify(context.Background(), mockRedisClient, "key", tc.args.err)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("OnModify() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

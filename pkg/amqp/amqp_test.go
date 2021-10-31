package amqp

import (
	"errors"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

func TestEnabled(t *testing.T) {
	cases := []struct {
		intention string
		instance  *Client
		want      bool
	}{
		{
			"empty",
			&Client{},
			false,
		},
		{
			"connection",
			&Client{
				connection: &amqp.Connection{},
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.Enabled(); got != tc.want {
				t.Errorf("Enabled() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	cases := []struct {
		intention string
		instance  *Client
		want      error
	}{
		{
			"empty",
			&Client{},
			nil,
		},
		{
			"not opened",
			&Client{},
			errors.New("amqp client closed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAMQPConnection := mocks.NewAMQPConnection(ctrl)

			switch tc.intention {
			case "not opened":
				tc.instance.connection = mockAMQPConnection
				mockAMQPConnection.EXPECT().IsClosed().Return(true)
			}

			got := tc.instance.Ping()

			failed := false

			if tc.want == nil && got != nil {
				failed = true
			} else if tc.want != nil && got == nil {
				failed = true
			} else if tc.want != nil && !strings.Contains(got.Error(), tc.want.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Ping() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

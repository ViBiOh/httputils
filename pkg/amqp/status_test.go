package amqp

import (
	"errors"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/mock/gomock"
)

func TestEnabled(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance *Client
		want     bool
	}{
		"empty": {
			&Client{},
			false,
		},
		"connection": {
			&Client{
				connection: &amqp.Connection{},
			},
			true,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.Enabled(); got != testCase.want {
				t.Errorf("Enabled() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance *Client
		want     error
	}{
		"empty": {
			&Client{},
			nil,
		},
		"not opened": {
			&Client{},
			errors.New("amqp client closed"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockAMQPConnection := mocks.NewAMQPConnection(ctrl)

			switch intention {
			case "not opened":
				testCase.instance.connection = mockAMQPConnection
				mockAMQPConnection.EXPECT().IsClosed().Return(true)
			}

			got := testCase.instance.Ping()

			failed := false

			if testCase.want == nil && got != nil {
				failed = true
			} else if testCase.want != nil && got == nil {
				failed = true
			} else if testCase.want != nil && !strings.Contains(got.Error(), testCase.want.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Ping() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func BenchmarkPing(b *testing.B) {
	instance := &Client{
		connection: &amqp.Connection{},
	}

	for b.Loop() {
		_ = instance.Ping()
	}
}

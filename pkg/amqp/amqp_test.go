package amqp

import (
	"errors"
	"strings"
	"testing"

	"github.com/streadway/amqp"
)

func TestEnabled(t *testing.T) {
	var cases = []struct {
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
	var cases = []struct {
		intention string
		instance  *Client
		want      error
	}{
		{
			"empty",
			&Client{},
			errors.New("amqp client disabled"),
		},
		{
			"not opened",
			&Client{
				connection: &amqp.Connection{},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
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

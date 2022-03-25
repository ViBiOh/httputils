package model

import (
	"testing"
	"time"
)

func TestSafeParseDuration(t *testing.T) {
	type args struct {
		name            string
		value           string
		defaultDuration time.Duration
	}

	cases := map[string]struct {
		args args
		want time.Duration
	}{
		"default": {
			args{
				name:            "test",
				value:           "abcd",
				defaultDuration: time.Minute,
			},
			time.Minute,
		},
		"parsed": {
			args{
				name:            "test",
				value:           "5m",
				defaultDuration: time.Minute,
			},
			time.Minute * 5,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := SafeParseDuration(tc.args.name, tc.args.value, tc.args.defaultDuration); got != tc.want {
				t.Errorf("SafeParseDuration() = %s, want %s", got, tc.want)
			}
		})
	}
}

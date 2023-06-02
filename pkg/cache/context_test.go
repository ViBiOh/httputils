package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBypassed(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		ctx  context.Context
		want bool
	}{
		"no value": {
			context.Background(),
			false,
		},
		"invalid type": {
			context.WithValue(context.Background(), bypassKey{}, "no"),
			false,
		},
		"false": {
			context.WithValue(context.Background(), bypassKey{}, false),
			false,
		},
		"function": {
			Bypass(context.Background()),
			true,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got := IsBypassed(testCase.ctx)

			assert.Equal(t, testCase.want, got)
		})
	}
}

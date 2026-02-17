package concurrent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChanUntilDone(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		source func(func()) <-chan int
	}

	cases := map[string]struct {
		args     args
		want     int
		wantDone int
	}{
		"simple with close": {
			args{
				context.Background(),
				func(cancel func()) <-chan int {
					ch := make(chan int, 5)

					go func() {
						defer close(ch)

						for i := range 3 {
							if i == 2 {
								cancel()
							}
							ch <- i
						}
					}()

					return ch
				},
			},
			3,
			1,
		},
		"simple no close": {
			args{
				context.Background(),
				func(cancel func()) <-chan int {
					ch := make(chan int, 5)

					go func() {
						for i := range 3 {
							if i == 2 {
								cancel()
							}
							ch <- i
						}
					}()

					return ch
				},
			},
			3,
			1,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())

			var actual int
			var actualDone int

			ChanUntilDone(ctx, testCase.args.source(cancel), func(input int) {
				actual += input
			}, func() {
				actualDone++
			})

			assert.Equal(t, testCase.want, actual)
			assert.Equal(t, testCase.wantDone, actualDone)
		})
	}
}

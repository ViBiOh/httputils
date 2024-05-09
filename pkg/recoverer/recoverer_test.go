package recoverer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()

		func() {
			defer Error(nil)

			panic("catch me if you can")
		}()
	})

	t.Run("valued", func(t *testing.T) {
		t.Parallel()

		var err error

		func() {
			defer Error(&err)

			panic("catch me if you can")
		}()

		assert.ErrorContains(t, err, "recovered")
	})

	t.Run("join", func(t *testing.T) {
		t.Parallel()

		err := errors.New("invalid")

		func() {
			defer Error(&err)

			panic("catch me if you can")
		}()

		assert.ErrorContains(t, err, "invalid")
		assert.ErrorContains(t, err, "recovered")
	})
}

func TestHandler(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{}{
		"simple": {},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			var err error

			func() {
				defer Handler(func(e error) {
					err = e
				})

				panic("catch me if you can")
			}()

			assert.ErrorContains(t, err, "recovered")
		})
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{}{
		"simple": {},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			func() {
				defer Logger()

				panic("catch me if you can")
			}()
		})
	}
}

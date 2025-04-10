package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		got := New()

		assert.Equal(t, 20, len(got))
	})
}

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		New()
	}
}

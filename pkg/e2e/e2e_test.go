package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2e(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := New("68991b71bcf849800b82e5e2862324f9")

		expected := []byte("Hello world! This is my secret message")

		got, gotErr := service.Encrypt(expected)
		assert.NoError(t, gotErr)

		got, gotErr = service.Decrypt(got)
		assert.NoError(t, gotErr)

		assert.Equal(t, expected, got)
	})
}

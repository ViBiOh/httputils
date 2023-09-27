package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string]()
		go instance.Start(ctx)

		instance.Set("hello", "world", time.Second)

		got, found := instance.Get("hello")

		expected := "world"

		assert.True(t, found)
		assert.Equal(t, expected, got)
	})

	t.Run("found but expired", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string]()
		go instance.Start(ctx)

		instance.Set("hello", "world", time.Millisecond*50)

		time.Sleep(time.Millisecond * 100)

		got, found := instance.Get("hello")

		expected := ""

		assert.False(t, found)
		assert.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string]()
		go instance.Start(ctx)

		got, found := instance.Get("hello")

		expected := ""

		assert.False(t, found)
		assert.Equal(t, expected, got)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("not present", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string]()
		go instance.Start(ctx)

		instance.Delete("hello")

		got, found := instance.Get("hello")

		expected := ""

		assert.False(t, found)
		assert.Equal(t, expected, got)
	})

	t.Run("present", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string]()
		go instance.Start(ctx)

		instance.Set("hello", "world", time.Second)
		instance.Delete("hello")

		time.Sleep(time.Millisecond * 100)

		got, found := instance.Get("hello")

		expected := ""

		assert.False(t, found)
		assert.Equal(t, expected, got)
	})
}

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

		instance := New[string, string](0)
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

		instance := New[string, string](0)
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

		instance := New[string, string](0)
		go instance.Start(ctx)

		got, found := instance.Get("hello")

		expected := ""

		assert.False(t, found)
		assert.Equal(t, expected, got)
	})
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	t.Run("nothing found", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)

		output := make([]string, 5)

		got := instance.GetAll([]string{"1", "2", "3", "4", "5"}, output)

		expected := make([]string, 5)
		expectedMissings := []string{"1", "2", "3", "4", "5"}

		assert.Equal(t, expected, output)
		assert.Equal(t, expectedMissings, got)
	})

	t.Run("part found", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)

		instance.Set("2", "two", 0)
		instance.Set("5", "five", 0)

		output := make([]string, 5)

		got := instance.GetAll([]string{"1", "2", "3", "4", "5"}, output)

		expected := []string{"", "two", "", "", "five"}
		expectedMissings := []string{"1", "3", "4"}

		assert.Equal(t, expected, output)
		assert.Equal(t, expectedMissings, got)
	})

	t.Run("all found", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)

		instance.Set("1", "one", 0)
		instance.Set("2", "two", 0)
		instance.Set("3", "three", 0)
		instance.Set("4", "four", 0)
		instance.Set("5", "five", 0)

		output := make([]string, 5)

		got := instance.GetAll([]string{"1", "2", "3", "4", "5"}, output)

		expected := []string{"one", "two", "three", "four", "five"}
		expectedMissings := []string(nil)

		assert.Equal(t, expected, output)
		assert.Equal(t, expectedMissings, got)
	})

	t.Run("lru", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](3)

		instance.Set("1", "one", 0)
		instance.Set("2", "two", 0)
		instance.Set("3", "three", 0)
		instance.Set("4", "four", 0)
		instance.Set("5", "five", 0)

		output := make([]string, 5)

		got := instance.GetAll([]string{"1", "2", "3", "4", "5"}, output)

		expected := []string{"", "", "three", "four", "five"}
		expectedMissings := []string{"1", "2"}

		assert.Equal(t, expected, output)
		assert.Equal(t, expectedMissings, got)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("not present", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		instance := New[string, string](0)
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

		instance := New[string, string](0)
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

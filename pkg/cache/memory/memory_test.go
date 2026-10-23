package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

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

		ctx := t.Context()

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

		ctx := t.Context()

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

		ctx := t.Context()

		instance := New[string, string](3)
		go instance.Start(ctx)

		instance.Set("1", "one", 0)
		instance.Set("2", "two", 0)
		instance.Set("3", "three", 0)
		instance.Set("4", "four", 0)
		instance.Set("5", "five", 0)

		time.Sleep(time.Second)

		output := make([]string, 5)

		got := instance.GetAll([]string{"1", "2", "3", "4", "5"}, output)

		expected := []string{"", "", "three", "four", "five"}
		expectedMissings := []string{"1", "2"}

		assert.Equal(t, expected, output)
		assert.Equal(t, expectedMissings, got)
	})
}

func BenchmarkLRURefresh(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			benchmarkLRURefresh(b, size)
		})
	}
}

func benchmarkLRURefresh(b *testing.B, size int) {
	b.Helper()

	instance := New[int, int](size)

	for i := range size {
		instance.Set(i, i, 0)
	}

	// Cycle through all elements so the scan position varies
	var i int

	for b.Loop() {
		instance.refresh(i % size)
		i++
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("not present", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

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

		ctx := t.Context()

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

func (cache *Cache[K, V]) size() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	return len(cache.content)
}

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("refresh extends ttl", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)

		instance.Set("hello", "world", time.Millisecond*50)
		instance.Set("hello", "again", time.Hour)

		time.Sleep(time.Millisecond * 150)

		got, found := instance.Get("hello")

		assert.True(t, found)
		assert.Equal(t, "again", got)
	})

	t.Run("zero ttl removes expiration", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)

		instance.Set("hello", "world", time.Millisecond*50)
		instance.Set("hello", "world", 0)

		time.Sleep(time.Millisecond * 150)

		_, found := instance.Get("hello")

		assert.True(t, found)
	})

	t.Run("without start", func(t *testing.T) {
		t.Parallel()

		instance := New[int, int](10)

		done := make(chan struct{})
		go func() {
			defer close(done)

			for i := range 1000 {
				instance.Set(i, i, time.Hour)
				instance.Delete(i - 5)
			}
		}()

		select {
		case <-done:
		case <-time.After(time.Second * 5):
			t.Fatal("Set blocked without a running Start")
		}

		assert.LessOrEqual(t, instance.size(), 10)
	})
}

func TestExpiration(t *testing.T) {
	t.Parallel()

	t.Run("evicted in background", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)
		go instance.Start(t.Context())

		instance.Set("hello", "world", time.Millisecond*50)

		assert.Eventually(t, func() bool { return instance.size() == 0 }, time.Second*2, time.Millisecond*10)
	})

	t.Run("delete of another key keeps the expiration", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)
		go instance.Start(t.Context())

		instance.Set("hello", "world", time.Millisecond*100)
		time.Sleep(time.Millisecond * 10)
		instance.Delete("unrelated")

		assert.Eventually(t, func() bool { return instance.size() == 0 }, time.Second*2, time.Millisecond*10)
	})

	t.Run("earlier expiration added after a later one", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)
		go instance.Start(t.Context())

		instance.Set("late", "world", time.Hour)
		time.Sleep(time.Millisecond * 10)
		instance.Set("soon", "world", time.Millisecond*50)

		assert.Eventually(t, func() bool { return instance.size() == 1 }, time.Second*2, time.Millisecond*10)

		_, found := instance.Get("late")
		assert.True(t, found)
	})

	t.Run("deleted entry is not evicted again", func(t *testing.T) {
		t.Parallel()

		instance := New[string, string](0)
		go instance.Start(t.Context())

		instance.Set("hello", "world", time.Millisecond*50)
		instance.Delete("hello")
		instance.Set("hello", "again", time.Hour)

		time.Sleep(time.Millisecond * 150)

		got, found := instance.Get("hello")

		assert.True(t, found)
		assert.Equal(t, "again", got)
	})
}

func TestStart(t *testing.T) {
	t.Parallel()

	t.Run("returns when context is cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		instance := New[string, string](10)

		done := make(chan struct{})
		go func() {
			defer close(done)

			instance.Start(ctx)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(time.Second * 5):
			t.Fatal("Start did not return after cancellation")
		}
	})

	t.Run("usable while and after cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		instance := New[int, int](10)
		go instance.Start(ctx)

		var wait sync.WaitGroup

		for worker := range 8 {
			wait.Go(func() {
				for i := range 500 {
					instance.Set(i%20, worker, time.Second)
					instance.Get(i % 20)
					instance.GetAll([]int{1, 2, 3}, make([]int, 3))
					instance.Delete((i + 7) % 20)
				}
			})
		}

		time.Sleep(time.Millisecond)
		cancel()

		wait.Wait()

		instance.Set(1, 1, time.Second)

		got, found := instance.Get(1)

		assert.True(t, found)
		assert.Equal(t, 1, got)
	})
}

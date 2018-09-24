package tools

import (
	"testing"
)

var entries = map[string]interface{}{`one`: `First`, `two`: `Second`, `three`: `Third`}

func initConccurentMapWithValues(values map[string]interface{}) *ConcurrentMap {
	concurrentMap := CreateConcurrentMap(5, 2)

	for key, content := range values {
		concurrentMap.Push(key, content)
	}

	return concurrentMap
}

func TestGet(t *testing.T) {
	var cases = []struct {
		entries map[string]interface{}
		ID      string
		want    interface{}
	}{
		{
			entries,
			`InvalidOne`,
			nil,
		},
		{
			entries,
			`one`,
			entries[`one`],
		},
	}

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)

		if result := concurrentMap.Get(testCase.ID); result != testCase.want {
			t.Errorf(`Get(%v) = %v, want %v`, testCase.ID, result, testCase.want)
		}

		concurrentMap.Close()
	}
}

func BenchmarkGet(b *testing.B) {
	var testCase = struct {
		entries map[string]interface{}
		key     string
	}{
		entries,
		`one`,
	}

	concurrentMap := initConccurentMapWithValues(testCase.entries)
	defer concurrentMap.Close()

	for i := 0; i < b.N; i++ {
		concurrentMap.Get(testCase.key)
	}
}

func TestPush(t *testing.T) {
	var cases = []struct {
		content interface{}
	}{
		{
			entries[`one`],
		},
	}

	for _, testCase := range cases {
		concurrentMap := CreateConcurrentMap(5, 2)
		concurrentMap.Push(`one`, testCase.content)

		if result := concurrentMap.Get(`one`); testCase.content != result {
			t.Errorf(`Push(%v) = %v, want %v`, testCase.content, result, testCase.content)
		}

		concurrentMap.Close()
	}
}

func BenchmarkPush(b *testing.B) {
	var testCase = struct {
		key   string
		value interface{}
	}{
		`one`,
		entries[`one`],
	}

	concurrentMap := CreateConcurrentMap(5, 2)
	defer concurrentMap.Close()

	for i := 0; i < b.N; i++ {
		concurrentMap.Push(testCase.key, testCase.value)
	}
}

func TestRemove(t *testing.T) {
	var cases = []struct {
		entries map[string]interface{}
		key     string
		want    bool
	}{
		{
			entries,
			`one`,
			true,
		},
		{
			entries,
			`Unknown`,
			false,
		},
	}

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)

		initial := concurrentMap.Get(testCase.key)
		concurrentMap.Remove(testCase.key)

		if result := concurrentMap.Get(testCase.key); (testCase.want && result == initial) || (!testCase.want && result != initial) {
			t.Errorf(`Remove(%v) = %v, want %v`, testCase.key, result, initial)
		}

		concurrentMap.Close()
	}
}

func BenchmarkRemove(b *testing.B) {
	var testCase = struct {
		key string
	}{
		`one`,
	}

	concurrentMap := CreateConcurrentMap(5, 2)
	defer concurrentMap.Close()

	for i := 0; i < b.N; i++ {
		concurrentMap.Remove(testCase.key)
	}
}

func TestList(t *testing.T) {
	var cases = []struct {
		entries map[string]interface{}
		want    int
	}{
		{
			entries,
			3,
		},
		{
			map[string]interface{}{},
			0,
		},
	}

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)

		result := 0
		for range concurrentMap.List() {
			result++
		}

		if result != testCase.want {
			t.Errorf(`List() = %v, want %v`, result, testCase.want)
		}

		concurrentMap.Close()
	}
}

func TestClose(t *testing.T) {
	var cases = []struct {
		entries map[string]interface{}
		want    int
	}{
		{
			entries,
			3,
		},
		{
			map[string]interface{}{},
			0,
		},
	}

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)

		if result := len(concurrentMap.Close()); result != testCase.want {
			t.Errorf(`Close() = %v, want %v`, result, testCase.want)
		}
	}
}

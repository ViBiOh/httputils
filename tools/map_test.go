package tools

import (
	"testing"
)

var entries = []MapContent{simpleMapContent{`First`}, simpleMapContent{`Second`}, simpleMapContent{`Third`}}

func initConccurentMapWithValues(values []MapContent) *ConcurrentMap {
	concurrentMap := CreateConcurrentMap(5, 2)

	for _, content := range values {
		concurrentMap.Push(content)
	}

	return concurrentMap
}

func TestGet(t *testing.T) {
	var tests = []struct {
		entries []MapContent
		ID      string
		want    MapContent
	}{
		{
			entries,
			`InvalidOne`,
			nil,
		},
		{
			entries,
			`Second`,
			entries[1],
		},
	}

	for _, test := range tests {
		concurrentMap := initConccurentMapWithValues(test.entries)
		defer concurrentMap.Close()

		if result := concurrentMap.Get(test.ID); result != test.want {
			t.Errorf(`Get(%v) = %v, want %v`, test.ID, result, test.want)
		}
	}
}

func TestPush(t *testing.T) {
	var tests = []struct {
		content MapContent
	}{
		{
			simpleMapContent{`Test`},
		},
	}

	for _, test := range tests {
		concurrentMap := CreateConcurrentMap(5, 2)
		defer concurrentMap.Close()

		concurrentMap.Push(test.content)

		if result := concurrentMap.Get(test.content.GetID()); test.content != result {
			t.Errorf(`Push(%v) = %v, want %v`, test.content, result, test.content)
		}
	}
}

func TestRemove(t *testing.T) {
	var tests = []struct {
		entries []MapContent
		ID      string
		want    bool
	}{
		{
			entries,
			`First`,
			true,
		},
		{
			entries,
			`Unknown`,
			false,
		},
	}

	for _, test := range tests {
		concurrentMap := initConccurentMapWithValues(test.entries)
		defer concurrentMap.Close()

		initial := concurrentMap.Get(test.ID)
		concurrentMap.Remove(test.ID)

		if result := concurrentMap.Get(test.ID); (test.want && result == initial) || (!test.want && result != initial) {
			t.Errorf(`Remove(%v) = %v, want %v`, test.ID, result, initial)
		}
	}
}

func TestList(t *testing.T) {
	var tests = []struct {
		entries []MapContent
		want    int
	}{
		{
			entries[:2],
			2,
		},
		{
			entries[2:3],
			1,
		},
		{
			[]MapContent{},
			0,
		},
	}

	for _, test := range tests {
		concurrentMap := initConccurentMapWithValues(test.entries)
		defer concurrentMap.Close()

		result := 0
		for range concurrentMap.List() {
			result++
		}

		if result != test.want {
			t.Errorf(`List() = %v, want %v`, result, test.want)
		}
	}
}

func TestClose(t *testing.T) {
	var tests = []struct {
		entries []MapContent
		want    int
	}{
		{
			entries[:2],
			2,
		},
		{
			entries[2:3],
			1,
		},
		{
			[]MapContent{},
			0,
		},
	}

	for _, test := range tests {
		concurrentMap := initConccurentMapWithValues(test.entries)

		if result := len(concurrentMap.Close()); result != test.want {
			t.Errorf(`Close() = %v, want %v`, result, test.want)
		}
	}
}

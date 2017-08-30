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
	var cases = []struct {
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

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)
		defer concurrentMap.Close()

		if result := concurrentMap.Get(testCase.ID); result != testCase.want {
			t.Errorf(`Get(%v) = %v, want %v`, testCase.ID, result, testCase.want)
		}
	}
}

func TestPush(t *testing.T) {
	var cases = []struct {
		content MapContent
	}{
		{
			simpleMapContent{`Test`},
		},
	}

	for _, testCase := range cases {
		concurrentMap := CreateConcurrentMap(5, 2)
		defer concurrentMap.Close()

		concurrentMap.Push(testCase.content)

		if result := concurrentMap.Get(testCase.content.GetID()); testCase.content != result {
			t.Errorf(`Push(%v) = %v, want %v`, testCase.content, result, testCase.content)
		}
	}
}

func TestRemove(t *testing.T) {
	var cases = []struct {
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

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)
		defer concurrentMap.Close()

		initial := concurrentMap.Get(testCase.ID)
		concurrentMap.Remove(testCase.ID)

		if result := concurrentMap.Get(testCase.ID); (testCase.want && result == initial) || (!testCase.want && result != initial) {
			t.Errorf(`Remove(%v) = %v, want %v`, testCase.ID, result, initial)
		}
	}
}

func TestList(t *testing.T) {
	var cases = []struct {
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

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)
		defer concurrentMap.Close()

		result := 0
		for range concurrentMap.List() {
			result++
		}

		if result != testCase.want {
			t.Errorf(`List() = %v, want %v`, result, testCase.want)
		}
	}
}

func TestClose(t *testing.T) {
	var cases = []struct {
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

	for _, testCase := range cases {
		concurrentMap := initConccurentMapWithValues(testCase.entries)

		if result := len(concurrentMap.Close()); result != testCase.want {
			t.Errorf(`Close() = %v, want %v`, result, testCase.want)
		}
	}
}

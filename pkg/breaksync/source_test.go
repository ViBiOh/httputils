package breaksync

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestComputeSynchro(t *testing.T) {
	t.Parallel()

	simple := NewSource(nil, Identity, nil)
	simple.currentKey = "AAAAA00000"

	substring := NewSource(nil, Identity, nil)
	substring.currentKey = "AAAAA00000"

	extrastring := NewSource(nil, Identity, nil)
	extrastring.currentKey = "AAAAA00000"

	unmatch := NewSource(nil, Identity, nil)
	unmatch.currentKey = "AAAAA00000"

	cases := map[string]struct {
		instance *Source[string]
		input    string
		want     bool
	}{
		"simple": {
			simple,
			"AAAAA00000",
			true,
		},
		"substring": {
			substring,
			"AAAAA",
			true,
		},
		"extrastring": {
			extrastring,
			"AAAAA00000zzzzz",
			true,
		},
		"unmatch": {
			unmatch,
			"AAAAA00001",
			false,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			testCase.instance.ComputeSynchro(testCase.input)
			if testCase.instance.synchronized != testCase.want {
				t.Errorf("computeSynchro() = %t, want %t", testCase.instance.synchronized, testCase.want)
			}
		})
	}
}

func TestSourceRead(t *testing.T) {
	t.Parallel()

	errRead := errors.New("read error")

	copyErr := NewSource(func() (string, error) {
		return "", errRead
	}, Identity, nil)
	copyErr.next = "Golang Test"
	copyErr.nextKey = "Golang"

	copyEnd := NewSource(func() (string, error) {
		return "", io.EOF
	}, Identity, nil)
	copyEnd.next = "Golang Test"
	copyEnd.nextKey = "Golang"

	cases := map[string]struct {
		instance       *Source[string]
		want           error
		wantCurrent    any
		wantCurrentKey string
		wantNext       any
		wantNextKey    string
	}{
		"copy next in current": {
			copyErr,
			errRead,
			"Golang Test",
			"Golang",
			"Golang Test",
			"Golang",
		},
		"error": {
			NewSource(func() (string, error) {
				return "", errRead
			}, Identity, nil),
			errRead,
			"",
			"",
			"",
			"",
		},
		"success": {
			NewSource(func() (string, error) {
				return "Gopher", io.EOF
			}, Identity, nil),
			nil,
			"",
			"",
			"Gopher",
			finalValue,
		},
		"end": {
			copyEnd,
			nil,
			"Golang Test",
			"Golang",
			"",
			finalValue,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			err := testCase.instance.read()

			if testCase.want != nil && !errors.Is(err, testCase.want) {
				t.Errorf("Read() = %v, want %v", err, testCase.want)
			} else if !reflect.DeepEqual(testCase.wantCurrent, testCase.instance.current) {
				t.Errorf("Read().Current = `%s`, want `%s`", testCase.wantCurrent, testCase.instance.current)
			} else if testCase.wantCurrentKey != testCase.instance.currentKey {
				t.Errorf("Read().currentKey = `%s`, want `%s`", testCase.wantCurrentKey, testCase.instance.currentKey)
			} else if !reflect.DeepEqual(testCase.wantNext, testCase.instance.next) {
				t.Errorf("Read().next = `%s`, want `%s`", testCase.wantNext, testCase.instance.next)
			} else if testCase.wantNextKey != testCase.instance.nextKey {
				t.Errorf("Read().nextKey = `%s`, want `%s`", testCase.wantNextKey, testCase.instance.nextKey)
			}
		})
	}
}

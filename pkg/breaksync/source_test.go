package breaksync

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestComputeSynchro(t *testing.T) {
	t.Parallel()

	simple := NewSource(nil, Identity, nil)
	simple.currentKey = []byte("AAAAA00000")

	substring := NewSource(nil, Identity, nil)
	substring.currentKey = []byte("AAAAA00000")

	extrastring := NewSource(nil, Identity, nil)
	extrastring.currentKey = []byte("AAAAA00000")

	unmatch := NewSource(nil, Identity, nil)
	unmatch.currentKey = []byte("AAAAA00000")

	cases := map[string]struct {
		instance *Source[string]
		input    []byte
		want     bool
	}{
		"simple": {
			simple,
			[]byte("AAAAA00000"),
			true,
		},
		"substring": {
			substring,
			[]byte("AAAAA"),
			true,
		},
		"extrastring": {
			extrastring,
			[]byte("AAAAA00000zzzzz"),
			true,
		},
		"unmatch": {
			unmatch,
			[]byte("AAAAA00001"),
			false,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

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
	copyErr.nextKey = []byte("Golang")

	copyEnd := NewSource(func() (string, error) {
		return "", io.EOF
	}, Identity, nil)
	copyEnd.next = "Golang Test"
	copyEnd.nextKey = []byte("Golang")

	cases := map[string]struct {
		instance       *Source[string]
		want           error
		wantCurrent    any
		wantCurrentKey []byte
		wantNext       any
		wantNextKey    []byte
	}{
		"copy next in current": {
			copyErr,
			errRead,
			"Golang Test",
			[]byte("Golang"),
			"Golang Test",
			[]byte("Golang"),
		},
		"error": {
			NewSource(func() (string, error) {
				return "", errRead
			}, Identity, nil),
			errRead,
			"",
			[]byte{},
			"",
			[]byte{},
		},
		"success": {
			NewSource(func() (string, error) {
				return "Gopher", io.EOF
			}, Identity, nil),
			nil,
			"",
			[]byte{},
			"Gopher",
			finalValue,
		},
		"end": {
			copyEnd,
			nil,
			"Golang Test",
			[]byte("Golang"),
			"",
			finalValue,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			err := testCase.instance.read()

			if testCase.want != nil && !errors.Is(err, testCase.want) {
				t.Errorf("Read() = %v, want %v", err, testCase.want)
			} else if !reflect.DeepEqual(testCase.wantCurrent, testCase.instance.current) {
				t.Errorf("Read().Current = `%s`, want `%s`", testCase.wantCurrent, testCase.instance.current)
			} else if !bytes.Equal(testCase.wantCurrentKey, testCase.instance.currentKey) {
				t.Errorf("Read().currentKey = `%s`, want `%s`", testCase.wantCurrentKey, testCase.instance.currentKey)
			} else if !reflect.DeepEqual(testCase.wantNext, testCase.instance.next) {
				t.Errorf("Read().next = `%s`, want `%s`", testCase.wantNext, testCase.instance.next)
			} else if !bytes.Equal(testCase.wantNextKey, testCase.instance.nextKey) {
				t.Errorf("Read().nextKey = `%s`, want `%s`", testCase.wantNextKey, testCase.instance.nextKey)
			}
		})
	}
}

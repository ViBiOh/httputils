package breaksync

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestComputeSynchro(t *testing.T) {
	simple := NewSource(nil, Identity, nil)
	simple.currentKey = "AAAAA00000"

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
			simple,
			"AAAAA",
			true,
		},
		"extrastring": {
			simple,
			"AAAAA00000zzzzz",
			true,
		},
		"unmatch": {
			simple,
			"AAAAA00001",
			false,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			tc.instance.ComputeSynchro(tc.input)
			if tc.instance.synchronized != tc.want {
				t.Errorf("computeSynchro() = %t, want %t", tc.instance.synchronized, tc.want)
			}
		})
	}
}

func TestSourceRead(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			err := tc.instance.read()

			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("Read() = %v, want %v", err, tc.want)
			} else if !reflect.DeepEqual(tc.wantCurrent, tc.instance.current) {
				t.Errorf("Read().Current = `%s`, want `%s`", tc.wantCurrent, tc.instance.current)
			} else if tc.wantCurrentKey != tc.instance.currentKey {
				t.Errorf("Read().currentKey = `%s`, want `%s`", tc.wantCurrentKey, tc.instance.currentKey)
			} else if !reflect.DeepEqual(tc.wantNext, tc.instance.next) {
				t.Errorf("Read().next = `%s`, want `%s`", tc.wantNext, tc.instance.next)
			} else if tc.wantNextKey != tc.instance.nextKey {
				t.Errorf("Read().nextKey = `%s`, want `%s`", tc.wantNextKey, tc.instance.nextKey)
			}
		})
	}
}

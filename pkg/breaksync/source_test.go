package breaksync

import (
	"errors"
	"reflect"
	"testing"
)

func TestComputeSynchro(t *testing.T) {
	simple := NewSource(nil, nil, nil)
	simple.currentKey = "AAAAA00000"

	var cases = []struct {
		intention string
		instance  *Source
		input     string
		want      bool
	}{
		{
			"simple",
			simple,
			"AAAAA00000",
			true,
		},
		{
			"substring",
			simple,
			"AAAAA",
			true,
		},
		{
			"extrastring",
			simple,
			"AAAAA00000zzzzz",
			true,
		},
		{
			"unmatch",
			simple,
			"AAAAA00001",
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			tc.instance.computeSynchro(tc.input)
			if tc.instance.synchronized != tc.want {
				t.Errorf("computeSynchro() = %t, want %t", tc.instance.synchronized, tc.want)
			}
		})
	}
}

func TestSourceRead(t *testing.T) {
	var errRead = errors.New("read error")

	copyErr := NewSource(func() (interface{}, error) {
		return nil, errRead
	}, sourceBasicKeyer, nil)
	copyErr.next = "Golang Test"
	copyErr.nextKey = "Golang"

	copyEnd := NewSource(func() (interface{}, error) {
		return nil, nil
	}, sourceBasicKeyer, nil)
	copyEnd.next = "Golang Test"
	copyEnd.nextKey = "Golang"

	var cases = []struct {
		intention      string
		instance       *Source
		want           error
		wantCurrent    interface{}
		wantCurrentKey string
		wantNext       interface{}
		wantNextKey    string
	}{
		{
			"copy next in current",
			copyErr,
			errRead,
			"Golang Test",
			"Golang",
			"Golang Test",
			"Golang",
		},
		{
			"error",
			NewSource(func() (interface{}, error) {
				return nil, errRead
			}, sourceBasicKeyer, nil),
			errRead,
			nil,
			"",
			nil,
			"",
		},
		{
			"success",
			NewSource(func() (interface{}, error) {
				return "Gopher", nil
			}, sourceBasicKeyer, nil),
			nil,
			nil,
			"",
			"Gopher",
			"\"Gopher\"",
		},
		{
			"end",
			copyEnd,
			nil,
			"Golang Test",
			"Golang",
			nil,
			finalValue,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
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

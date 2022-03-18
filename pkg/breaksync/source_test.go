package breaksync

import (
	"errors"
	"reflect"
	"testing"
)

type identifiableString string

func (is identifiableString) Key() string {
	return string(is)
}

func TestComputeSynchro(t *testing.T) {
	simple := NewSource(nil, nil)
	simple.currentKey = "AAAAA00000"

	cases := []struct {
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
	errRead := errors.New("read error")

	copyErr := NewSource(func() (Identifiable, error) {
		return nil, errRead
	}, nil)
	copyErr.next = identifiableString("Golang Test")
	copyErr.nextKey = "Golang"

	copyEnd := NewSource(func() (Identifiable, error) {
		return nil, nil
	}, nil)
	copyEnd.next = identifiableString("Golang Test")
	copyEnd.nextKey = "Golang"

	cases := []struct {
		intention      string
		instance       *Source
		want           error
		wantCurrent    any
		wantCurrentKey string
		wantNext       any
		wantNextKey    string
	}{
		{
			"copy next in current",
			copyErr,
			errRead,
			identifiableString("Golang Test"),
			"Golang",
			identifiableString("Golang Test"),
			"Golang",
		},
		{
			"error",
			NewSource(func() (Identifiable, error) {
				return nil, errRead
			}, nil),
			errRead,
			nil,
			"",
			nil,
			"",
		},
		{
			"success",
			NewSource(func() (Identifiable, error) {
				return identifiableString("Gopher"), nil
			}, nil),
			nil,
			nil,
			"",
			identifiableString("Gopher"),
			"Gopher",
		},
		{
			"end",
			copyEnd,
			nil,
			identifiableString("Golang Test"),
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

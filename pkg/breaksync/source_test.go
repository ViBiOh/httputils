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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			testCase.instance.computeSynchro(testCase.input)
			if testCase.instance.synchronized != testCase.want {
				t.Errorf("computeSynchro() = %t, want %t", testCase.instance.synchronized, testCase.want)
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			err := testCase.instance.read()

			if testCase.want != nil && !errors.Is(err, testCase.want) {
				t.Errorf("Read() = %v, want %v", err, testCase.want)
			} else if !reflect.DeepEqual(testCase.wantCurrent, testCase.instance.Current) {
				t.Errorf("Read().Current =`%s`, want`%s`", testCase.wantCurrent, testCase.instance.Current)
			} else if testCase.wantCurrentKey != testCase.instance.currentKey {
				t.Errorf("Read().currentKey =`%s`, want`%s`", testCase.wantCurrentKey, testCase.instance.currentKey)
			} else if !reflect.DeepEqual(testCase.wantNext, testCase.instance.next) {
				t.Errorf("Read().next =`%s`, want`%s`", testCase.wantNext, testCase.instance.next)
			} else if testCase.wantNextKey != testCase.instance.nextKey {
				t.Errorf("Read().nextKey =`%s`, want`%s`", testCase.wantNextKey, testCase.instance.nextKey)
			}
		})
	}
}

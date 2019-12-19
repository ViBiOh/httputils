package breaksync

import (
	"errors"
	"fmt"
	"testing"
)

type client struct {
	name string
	card string
}

func TestRun(t *testing.T) {
	cards := []interface{}{
		"MASTERCARD",
		"VISA",
	}
	cardKeyer := func(o interface{}) string {
		return fmt.Sprintf("%-10s", o)
	}

	clients := []interface{}{
		client{"Bob", "MASTERCARD"},
		client{"Chuck", "MASTERCARD"},
		client{"Hulk", "MASTERCARD"},
		client{"Hulk", "MASTERCARD"},
		client{"Luke", "MASTERCARD"},
		client{"Superman", "MASTERCARD"},
		client{"Tony Stark", "MASTERCARD"},
		client{"Vador", "MASTERCARD"},
		client{"Yoda", "MASTERCARD"},
		client{"Einstein", "VISA"},
		client{"Vincent", "VISA"},
	}
	clientKeyer := func(o interface{}) string {
		c := o.(client)
		return fmt.Sprintf("%-10s%s", c.card, c.name)
	}

	cardRupture := NewRupture("card", func(i string) string {
		return fmt.Sprintf("%.10s", i)
	})

	errRead := errors.New("test error")
	numberReader := func(start int, failure bool) func() (interface{}, error) {
		i := start

		return func() (interface{}, error) {
			i++

			if i < 0 {
				return 0, errRead
			}

			if i <= 5 {
				return i, nil
			}

			if failure {
				return 0, errRead
			}

			return nil, nil
		}
	}

	var cases = []struct {
		intention    string
		sources      []*Source
		ruptures     []*Rupture
		businessFail bool
		want         int
		wantErr      error
	}{
		{
			"fully synchronized",
			[]*Source{
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
			},
			nil,
			false,
			5,
			nil,
		},
		{
			"desynchronized once",
			[]*Source{
				NewSource(numberReader(1, false), sourceBasicKeyer, nil),
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
			},
			nil,
			false,
			4,
			nil,
		},
		{
			"read first error",
			[]*Source{
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
				NewSource(numberReader(-2, false), sourceBasicKeyer, nil),
			},
			nil,
			false,
			0,
			errRead,
		},
		{
			"read later error",
			[]*Source{
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
				NewSource(numberReader(0, true), sourceBasicKeyer, nil),
			},
			nil,
			false,
			4,
			errRead,
		},
		{
			"business error",
			[]*Source{
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
				NewSource(numberReader(0, false), sourceBasicKeyer, nil),
			},
			nil,
			true,
			4,
			errRead,
		},
		{
			"should work with basic rupture on read",
			[]*Source{
				newSliceSource(clients, clientKeyer, nil),
				newSliceSource(cards, cardKeyer, cardRupture),
			},
			[]*Rupture{cardRupture},
			false,
			11,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			synchronization := NewSynchronization(testCase.sources, testCase.ruptures)

			var result int
			err := synchronization.Run(func(s *Synchronization) error {
				for _, source := range s.Sources {
					if !source.synchronized {
						return nil
					}
				}

				if result > 3 && testCase.businessFail {
					return errRead
				}

				result++
				return nil
			})

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				t.Errorf("Read() = %v, want %v", err, testCase.wantErr)
			} else if testCase.want != result {
				t.Errorf("Run() = %d, want %d", result, testCase.want)
			}
		})
	}
}

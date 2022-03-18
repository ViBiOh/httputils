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
	cards := []any{
		"AMEX",
		"MASTERCARD",
		"VISA",
		"WESTERN",
	}
	cardKeyer := func(o any) string {
		return fmt.Sprintf("%-10s", o)
	}

	clients := []any{
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
	clientKeyer := func(o any) string {
		c := o.(client)
		return fmt.Sprintf("%-10s%s", c.card, c.name)
	}

	cardRupture := NewRupture("card", func(i string) string {
		return fmt.Sprintf("%.10s", i)
	})

	errRead := errors.New("test error")
	numberReader := func(start int, failure bool) func() (any, error) {
		i := start

		return func() (any, error) {
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

	cases := []struct {
		intention    string
		instance     *Synchronization
		businessFail bool
		want         int
		wantErr      error
	}{
		{
			"fully synchronized",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), sourceBasicKeyer, nil), NewSource(numberReader(0, false), sourceBasicKeyer, nil)),
			false,
			5,
			nil,
		},
		{
			"desynchronized once",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), sourceBasicKeyer, nil), NewSource(numberReader(1, false), sourceBasicKeyer, nil)),
			false,
			4,
			nil,
		},
		{
			"read first error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), sourceBasicKeyer, nil), NewSource(numberReader(-2, false), sourceBasicKeyer, nil)),
			false,
			0,
			errRead,
		},
		{
			"read later error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), sourceBasicKeyer, nil), NewSource(numberReader(0, true), sourceBasicKeyer, nil)),
			false,
			4,
			errRead,
		},
		{
			"business error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), sourceBasicKeyer, nil), NewSource(numberReader(0, false), sourceBasicKeyer, nil)),
			true,
			4,
			errRead,
		},
		{
			"should work with basic rupture on read",
			NewSynchronization().AddSources(NewSliceSource(clients, clientKeyer, nil), NewSliceSource(cards, cardKeyer, cardRupture)).AddRuptures(cardRupture),
			false,
			11,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			var result int
			err := tc.instance.Run(func(synchronization uint64, items []any) error {
				if synchronization != 0 {
					return nil
				}

				if result > 3 && tc.businessFail {
					return errRead
				}

				result++
				return nil
			})

			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("Read() = %v, want %v", err, tc.wantErr)
			} else if tc.want != result {
				t.Errorf("Run() = %d, want %d", result, tc.want)
			}
		})
	}
}

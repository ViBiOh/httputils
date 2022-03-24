package breaksync

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"
)

type card string

func cardKeyer(c card) string {
	return string(c)
}

func numberKeyer(n int) string {
	return strconv.Itoa(n)
}

type client struct {
	name string
	card string
}

func clientKeyer(c client) string {
	return c.card
}

func TestRun(t *testing.T) {
	cards := []card{
		"AMEX",
		"MASTERCARD",
		"VISA",
		"WESTERN",
	}

	cardReader := make(chan card, 1)

	go func() {
		defer close(cardReader)

		for _, card := range cards {
			cardReader <- card
		}
	}()

	clients := []client{
		{"Bob", "MASTERCARD"},
		{"Chuck", "MASTERCARD"},
		{"Hulk", "MASTERCARD"},
		{"Hulk", "MASTERCARD"},
		{"Luke", "MASTERCARD"},
		{"Superman", "MASTERCARD"},
		{"Tony Stark", "MASTERCARD"},
		{"Vador", "MASTERCARD"},
		{"Yoda", "MASTERCARD"},
		{"Einstein", "VISA"},
		{"Vincent", "VISA"},
	}

	cardRupture := NewRupture("card", func(i string) string {
		return fmt.Sprintf("%.10s", i)
	})

	errRead := errors.New("test error")
	numberReader := func(start int, failure bool) func() (int, error) {
		i := start

		return func() (int, error) {
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

			return 0, io.EOF
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
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, false), numberKeyer, nil)),
			false,
			5,
			nil,
		},
		{
			"desynchronized once",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(1, false), numberKeyer, nil)),
			false,
			4,
			nil,
		},
		{
			"read first error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(-2, false), numberKeyer, nil)),
			false,
			0,
			errRead,
		},
		{
			"read later error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, true), numberKeyer, nil)),
			false,
			4,
			errRead,
		},
		{
			"business error",
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, false), numberKeyer, nil)),
			true,
			4,
			errRead,
		},
		{
			"should work with basic rupture on read",
			NewSynchronization().AddSources(NewSliceSource(clients, clientKeyer, nil), NewChanSource(cardReader, cardKeyer, cardRupture)),
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

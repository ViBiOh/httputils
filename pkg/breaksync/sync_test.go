package breaksync

import (
	"errors"
	"io"
	"strconv"
	"testing"
)

type card string

func cardKeyer(c card) []byte {
	return []byte(c)
}

func numberKeyer(n int) []byte {
	return []byte(strconv.Itoa(n))
}

type client struct {
	name string
	card string
}

func clientKeyer(c client) []byte {
	return []byte(c.card)
}

func TestRun(t *testing.T) {
	t.Parallel()

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

	cardRupture := NewRupture("card", func(i []byte) []byte {
		output := make([]byte, 10)

		copy(output[10-len(i):], i)
		return output
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

	cases := map[string]struct {
		instance     *Synchronization
		businessFail bool
		want         int
		wantErr      error
	}{
		"fully synchronized": {
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, false), numberKeyer, nil)),
			false,
			5,
			nil,
		},
		"desynchronized once": {
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(1, false), numberKeyer, nil)),
			false,
			4,
			nil,
		},
		"read first error": {
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(-2, false), numberKeyer, nil)),
			false,
			0,
			errRead,
		},
		"read later error": {
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, true), numberKeyer, nil)),
			false,
			4,
			errRead,
		},
		"business error": {
			NewSynchronization().AddSources(NewSource(numberReader(0, false), numberKeyer, nil), NewSource(numberReader(0, false), numberKeyer, nil)),
			true,
			4,
			errRead,
		},
		"should work with basic rupture on read": {
			NewSynchronization().AddSources(NewSliceSource(clients, clientKeyer, nil), NewChanSource(cardReader, cardKeyer, cardRupture)),
			false,
			11,
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			var result int
			err := testCase.instance.Run(func(synchronization uint64, items []any) error {
				if synchronization != 0 {
					return nil
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

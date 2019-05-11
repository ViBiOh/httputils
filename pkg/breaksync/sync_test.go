package breaksync

import (
	"fmt"
	"testing"
)

func TestAlgorithm(t *testing.T) {
	numberReader := func(start int) func() (interface{}, error) {
		i := start

		return func() (interface{}, error) {
			i = i + 2
			if i < 10 {
				return i, nil
			}
			return nil, nil
		}
	}

	var cases = []struct {
		intention string
		sources   []*Source
		ruptures  []*Rupture
		want      int
	}{
		{
			"should work two list fully synchronized",
			[]*Source{
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
			},
			[]*Rupture{},
			40,
		},
		{
			"should work two list never synchronized at the same time",
			[]*Source{
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
				NewSource(numberReader(-1), SourceBasicKeyer, nil),
			},
			[]*Rupture{},
			45,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			synchronization := NewSynchronization(testCase.sources, testCase.ruptures)

			var result int
			synchronization.Run(func(s *Synchronization) {
				for _, source := range s.Sources {
					if source.synchronized {
						result = result + source.Current.(int)
					}
				}
			})

			if testCase.want != result {
				t.Errorf("BreakSync Algorithm(%v) = %v, want %v", testCase.sources, result, testCase.want)
			}
		})
	}
}

type client struct {
	name string
	card string
}

func TestAlgorithmWithRupture(t *testing.T) {
	cards := []string{"MASTERCARD", "VISA"}
	cardKeyer := func(o interface{}) string {
		return fmt.Sprintf("%-10s", o)
	}
	cardRupture := NewRupture("card", func(i string) string {
		return fmt.Sprintf("%.10s", i)
	})

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
	clientKeyer := func(o interface{}) string {
		c := o.(client)

		return fmt.Sprintf("%-10s%s", c.card, c.name)
	}

	interfaceCards := make([]interface{}, len(cards))
	for i, d := range cards {
		interfaceCards[i] = d
	}

	interfaceClients := make([]interface{}, len(clients))
	for i, d := range clients {
		interfaceClients[i] = d
	}

	var cases = []struct {
		intention string
		sources   []*Source
		ruptures  []*Rupture
		want      uint
	}{
		{
			"should work with basic rupture on read",
			[]*Source{
				NewSliceSource(interfaceClients, clientKeyer, nil),
				NewSliceSource(interfaceCards, cardKeyer, cardRupture),
			},
			[]*Rupture{cardRupture},
			11,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			synchronization := NewSynchronization(testCase.sources, testCase.ruptures)

			result := uint(0)
			synchronization.Run(func(s *Synchronization) {
				allSynchronized := true
				for _, source := range s.Sources {
					if !source.synchronized {
						allSynchronized = false
					}
				}

				if allSynchronized {
					result++
				}
			})

			if testCase.want != result {
				t.Errorf("BreakSync Algorithm(%v) = %v, want %v", testCase.sources, result, testCase.want)
			}
		})
	}
}

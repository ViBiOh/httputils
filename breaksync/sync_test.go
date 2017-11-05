package breaksync

import (
	"fmt"
	"testing"
)

func Test_Algorithm(t *testing.T) {
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
			`should work two list fully synchronized`,
			[]*Source{
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
			},
			[]*Rupture{},
			40,
		},
		{
			`should work two list never synchronized at the same time`,
			[]*Source{
				NewSource(numberReader(-2), SourceBasicKeyer, nil),
				NewSource(numberReader(-1), SourceBasicKeyer, nil),
			},
			[]*Rupture{},
			45,
		},
	}

	for _, testCase := range cases {
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
			t.Errorf("%v\nBreakSync Algorithm(%v) = %v, want %v", testCase.intention, testCase.sources, result, testCase.want)
		}
	}
}

type client struct {
	name string
	card string
}

func Test_AlgorithmWithRupture(t *testing.T) {
	cardIndex := -1
	cards := []string{`MASTERCARD`, `VISA`}
	cardsReader := func() (interface{}, error) {
		cardIndex++
		if cardIndex < len(cards) {
			return cards[cardIndex], nil
		}
		return nil, nil
	}
	cardKeyer := func(o interface{}) string {
		return fmt.Sprintf(`%-10s`, o)
	}
	cardRupture := NewRupture(`card`, func(i string) string {
		return fmt.Sprintf(`%.10s`, i)
	})

	clientIndex := -1
	clients := []client{
		{`Bob`, `MASTERCARD`},
		{`Chuck`, `MASTERCARD`},
		{`Hulk`, `MASTERCARD`},
		{`Hulk`, `MASTERCARD`},
		{`Luke`, `MASTERCARD`},
		{`Superman`, `MASTERCARD`},
		{`Tony Stark`, `MASTERCARD`},
		{`Vador`, `MASTERCARD`},
		{`Yoda`, `MASTERCARD`},
		{`Einstein`, `VISA`},
		{`Vincent`, `VISA`},
	}
	clientsReader := func() (interface{}, error) {
		clientIndex++
		if clientIndex < len(clients) {
			return clients[clientIndex], nil
		}
		return nil, nil
	}
	clientKeyer := func(o interface{}) string {
		c := o.(client)

		return fmt.Sprintf(`%-10s%s`, c.card, c.name)
	}

	var cases = []struct {
		intention string
		sources   []*Source
		ruptures  []*Rupture
		want      int
	}{
		{
			`should work with basic rupture on read`,
			[]*Source{
				NewSource(clientsReader, clientKeyer, nil),
				NewSource(cardsReader, cardKeyer, cardRupture),
			},
			[]*Rupture{cardRupture},
			11,
		},
	}

	for _, testCase := range cases {
		synchronization := NewSynchronization(testCase.sources, testCase.ruptures)

		result := 0
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
			t.Errorf("%v\nBreakSync Algorithm(%v) = %v, want %v", testCase.intention, testCase.sources, result, testCase.want)
		}
	}
}

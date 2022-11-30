package breaksync

import "bytes"

var finalValue = []byte{0xFF, 0xFF, 0xFF, 0xFF}

type Synchronization struct {
	currentKey []byte
	nextKey    []byte

	sources  []SyncSource
	ruptures []*Rupture

	end bool
}

func NewSynchronization() *Synchronization {
	return &Synchronization{
		end: false,
	}
}

func (s *Synchronization) AddSources(sources ...SyncSource) *Synchronization {
	s.sources = append(s.sources, sources...)

	for _, source := range sources {
		if readRupture := source.ReadRupture(); readRupture != nil {
			s.ruptures = append(s.ruptures, readRupture)
		}
	}

	return s
}

func (s *Synchronization) Run(business func(uint64, []any) error) (err error) {
	if err = s.read(); err != nil {
		return
	}
	s.computeKey()

	items := make([]any, len(s.sources))

	for !s.end {
		if err = s.read(); err != nil {
			return
		}

		s.computeSynchro()
		s.computeKey()
		s.computeRuptures()

		if err = business(s.computeItems(items), items); err != nil {
			return
		}
	}

	return nil
}

func (s *Synchronization) read() error {
	for _, source := range s.sources {
		if err := source.Read(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Synchronization) computeKey() {
	s.currentKey = s.nextKey
	s.nextKey = finalValue

	for _, source := range s.sources {
		if source.IsSynchronized() {
			if nextKey := source.NextKey(); bytes.Compare(nextKey, s.nextKey) < 0 {
				s.nextKey = nextKey
			}
		} else if currentKey := source.CurrentKey(); bytes.Compare(currentKey, s.nextKey) < 0 {
			s.nextKey = currentKey
		}
	}

	s.end = bytes.Equal(s.nextKey, finalValue)
}

func (s *Synchronization) computeSynchro() {
	for _, source := range s.sources {
		source.ComputeSynchro(s.nextKey)
	}
}

func (s *Synchronization) computeRuptures() {
	init := false

	for _, rupture := range s.ruptures {
		init = rupture.compute(s.currentKey, s.nextKey, init)
	}
}

func (s *Synchronization) computeItems(items []any) uint64 {
	var itemsFlags uint64

	for i, source := range s.sources {
		if !source.IsSynchronized() {
			itemsFlags += 1 << i
		}

		items[i] = source.Current()
	}

	return itemsFlags
}

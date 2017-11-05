package breaksync

import "fmt"

const finalValue = "\uffff"

// Synchronization in a break/sync algorithm
type Synchronization struct {
	currentKey string
	nextKey    string
	end        bool

	Sources  []*Source
	ruptures []*Rupture
}

// NewSynchronization creates and initializes Synchronization
func NewSynchronization(sources []*Source, ruptures []*Rupture) *Synchronization {
	return &Synchronization{Sources: sources, ruptures: ruptures, end: false}
}

func (s *Synchronization) read() error {
	for _, source := range s.Sources {
		if source.synchronized && (source.readRupture == nil || source.readRupture.last) {
			if _, err := source.read(); err != nil {
				return fmt.Errorf(`Error while reading source: %v`, err)
			}
		}
	}

	return nil
}

func (s *Synchronization) computeKeys() {
	s.currentKey = s.nextKey

	s.nextKey = finalValue
	for _, source := range s.Sources {
		if source.synchronized {
			if source.nextKey < s.nextKey {
				s.nextKey = source.nextKey
			}
		} else if source.currentKey < s.nextKey {
			s.nextKey = source.currentKey
		}
	}

	s.end = s.nextKey == finalValue
}

func (s *Synchronization) computeSynchro() {
	for _, source := range s.Sources {
		source.computeSynchro(s.nextKey)
	}
}

func (s *Synchronization) computeRuptures() {
	init := false

	for _, rupture := range s.ruptures {
		init = rupture.compute(s.currentKey, s.nextKey, init)
	}
}

func (s *Synchronization) computeSynchros() {
	for _, source := range s.Sources {
		source.computeSynchro(s.currentKey)
	}
}

// Run start break/sync algorithm
func (s *Synchronization) Run(business func(*Synchronization)) {
	s.read()
	s.computeKeys()

	for !s.end {
		s.read()
		s.computeSynchro()
		s.computeKeys()
		s.computeRuptures()

		business(s)
	}
}

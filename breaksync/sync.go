package breaksync

const finalValue = "\uffff"

type Synchronization struct {
	currentKey string
	nextKey    string
	end        bool

	sources  []*Source
	ruptures []*Rupture
}

func (s *Synchronization) computeKeys() {
	s.currentKey = s.nextKey

	s.nextKey = finalValue
	for _, source := range s.sources {
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

func (s *Synchronization) computeRuptures() {
	init := false

	for _, rupture := range s.ruptures {
		init = rupture.compute(s.currentKey, s.nextKey, init)
	}
}

func (s *Synchronization) computeSynchros() {
	for _, source := range s.sources {
		source.computeSynchro(s.currentKey)
	}
}

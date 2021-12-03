package breaksync

const finalValue = "\uffff"

// Synchronization in a break/sync algorithm
type Synchronization struct {
	currentKey string
	nextKey    string

	sources  []*Source
	ruptures []*Rupture

	end bool
}

// NewSynchronization creates and initializes Synchronization
func NewSynchronization() *Synchronization {
	return &Synchronization{
		end: false,
	}
}

// AddSources adds given source
func (s *Synchronization) AddSources(sources ...*Source) *Synchronization {
	s.sources = append(s.sources, sources...)

	return s
}

// AddRuptures adds given rupture
func (s *Synchronization) AddRuptures(ruptures ...*Rupture) *Synchronization {
	s.ruptures = append(s.ruptures, ruptures...)

	return s
}

// Run start break/sync algorithm
func (s *Synchronization) Run(business func(uint64, []interface{}) error) (err error) {
	if err = s.read(); err != nil {
		return
	}
	s.computeKey()

	items := make([]interface{}, len(s.sources))

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
		if !source.synchronized {
			continue
		}

		if source.readRupture != nil && !source.readRupture.last {
			continue
		}

		if err := source.read(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Synchronization) computeKey() {
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

func (s *Synchronization) computeSynchro() {
	for _, source := range s.sources {
		source.computeSynchro(s.nextKey)
	}
}

func (s *Synchronization) computeRuptures() {
	init := false

	for _, rupture := range s.ruptures {
		init = rupture.compute(s.currentKey, s.nextKey, init)
	}
}

func (s *Synchronization) computeItems(items []interface{}) uint64 {
	var itemsFlags uint64

	for i, source := range s.sources {
		if !source.synchronized {
			itemsFlags += 1 << i
		}
		items[i] = source.current
	}

	return itemsFlags
}

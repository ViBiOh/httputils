package breaksync

const finalValue = "\uffff"

// Synchronization in a break/sync algorithm.
type Synchronization struct {
	currentKey string
	nextKey    string

	sources  []SyncSource
	ruptures []*Rupture

	end bool
}

// NewSynchronization creates and initializes Synchronization.
func NewSynchronization() *Synchronization {
	return &Synchronization{
		end: false,
	}
}

// AddSources adds given source.
func (s *Synchronization) AddSources(sources ...SyncSource) *Synchronization {
	s.sources = append(s.sources, sources...)

	for _, source := range sources {
		if readRupture := source.ReadRupture(); readRupture != nil {
			s.ruptures = append(s.ruptures, readRupture)
		}
	}

	return s
}

// Run start break/sync algorithm.
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
			if nextKey := source.NextKey(); nextKey < s.nextKey {
				s.nextKey = nextKey
			}
		} else if currentKey := source.CurrentKey(); currentKey < s.nextKey {
			s.nextKey = currentKey
		}
	}

	s.end = s.nextKey == finalValue
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

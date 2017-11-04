package breaksync

// Source of data in a break/sync algorithm
type Source struct {
	synchronized bool
	reader       func() (interface{}, error)
	keyer        func(interface{}) string
	readRupture  *Rupture

	current    interface{}
	currentKey string
	next       interface{}
	nextKey    string
}

// NewSource creates and initialize Source
func NewSource(reader func() (interface{}, error), keyer func(interface{}) string, readRupture *Rupture) *Source {
	return &Source{synchronized: true, reader: reader, keyer: keyer, readRupture: readRupture}
}

func (s *Source) computeSynchro(key string) {
	s.synchronized = key == s.currentKey
}

func (s *Source) read() (interface{}, error) {
	s.current = s.next
	s.currentKey = s.nextKey

	next, err := s.reader()
	if err == nil {
		s.next = next
		if next != nil {
			s.nextKey = s.keyer(next)
		} else {
			s.nextKey = finalValue
		}
	}

	return next, err
}

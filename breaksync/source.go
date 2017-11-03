package breaksync

type Source struct {
	synchronized bool

	current    interface{}
	currentKey string
	next       interface{}
	nextKey    string

	reader func() (interface{}, error)
	keyer  func(interface{}) string
}

func (s *Source) Read() (interface{}, error) {
	s.current = s.next
	s.currentKey = s.nextKey

	next, err := s.reader()
	if err != nil {
		return nil, err
	}

	if next != nil {
		s.next = next
		s.nextKey = s.keyer(next)
	} else {
		s.next = nil
		s.nextKey = finalValue
	}

	return next, nil
}

func (s *Source) computeSynchro(key string) {
	s.synchronized = key == s.currentKey
}

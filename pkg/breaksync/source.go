package breaksync

// Source of data in a break/sync algorithm
type Source struct {
	next    any
	current any

	reader func() (any, error)
	keyer  func(any) string

	readRupture *Rupture

	currentKey string
	nextKey    string

	synchronized bool
}

// NewSource creates and initialize Source
func NewSource(reader func() (any, error), keyer func(any) string, readRupture *Rupture) *Source {
	return &Source{
		synchronized: true,
		reader:       reader,
		keyer:        keyer,
		readRupture:  readRupture,
	}
}

func (s *Source) computeSynchro(key string) {
	compareKey := s.currentKey[:min(len(key), len(s.currentKey))]
	s.synchronized = compareKey == key[:len(compareKey)]
}

func (s *Source) read() error {
	s.current = s.next
	s.currentKey = s.nextKey

	next, err := s.reader()
	if err != nil {
		return err
	}

	s.next = next
	if next != nil {
		s.nextKey = s.keyer(next)
	} else {
		s.nextKey = finalValue
	}

	return nil
}

// NewSliceSource is a source from a slice, read sequentially
func NewSliceSource(arr []any, keyer func(any) string, readRupture *Rupture) *Source {
	index := -1

	return NewSource(func() (any, error) {
		index++
		if index < len(arr) {
			return arr[index], nil
		}
		return nil, nil
	}, keyer, readRupture)
}

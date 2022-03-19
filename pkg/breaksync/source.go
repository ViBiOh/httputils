package breaksync

// IdentityRupture consider key as rupture value
var IdentityRupture = func(a string) string {
	return a
}

// Identifiable is an object can that provide identification
type Identifiable interface {
	Key() string
}

// Source of data in a break/sync algorithm
type Source struct {
	next    Identifiable
	current Identifiable

	reader func() (Identifiable, error)

	readRupture *Rupture

	currentKey string
	nextKey    string

	synchronized bool
}

// NewSource creates and initialize Source
func NewSource(reader func() (Identifiable, error), readRupture *Rupture) *Source {
	return &Source{
		synchronized: true,
		reader:       reader,
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
		s.nextKey = s.next.Key()
	} else {
		s.nextKey = finalValue
	}

	return nil
}

// NewSliceSource is a source from a slice, read sequentially
func NewSliceSource[T Identifiable](arr []T, readRupture *Rupture) *Source {
	index := -1

	return NewSource(func() (Identifiable, error) {
		index++
		if index < len(arr) {
			return arr[index], nil
		}
		return nil, nil
	}, readRupture)
}

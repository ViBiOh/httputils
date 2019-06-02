package breaksync

import (
	"fmt"
	"strconv"
)

// Source of data in a break/sync algorithm
type Source struct {
	synchronized bool
	reader       func() (interface{}, error)
	keyer        func(interface{}) string
	readRupture  *Rupture

	Current    interface{}
	currentKey string
	next       interface{}
	nextKey    string
}

// NewSource creates and initialize Source
func NewSource(reader func() (interface{}, error), keyer func(interface{}) string, readRupture *Rupture) *Source {
	return &Source{synchronized: true, reader: reader, keyer: keyer, readRupture: readRupture}
}

// NewSliceSource creates source from given slice
func NewSliceSource(slice []interface{}, keyer func(interface{}) string, readRupture *Rupture) *Source {
	index := -1

	return NewSource(func() (interface{}, error) {
		index++
		if index < len(slice) {
			return slice[index], nil
		}
		return nil, nil
	}, keyer, readRupture)
}

// SourceBasicKeyer basic keyer for string conversion
func SourceBasicKeyer(e interface{}) string {
	return fmt.Sprintf("%#v", e)
}

func (s *Source) computeSynchro(key string) {
	s.synchronized = fmt.Sprintf("%."+strconv.Itoa(len(s.currentKey))+"s", key) == s.currentKey
}

func (s *Source) read() (interface{}, error) {
	s.Current = s.next
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

package breaksync

import (
	"io"
)

var Identity = func(a string) string {
	return a
}

var _ SyncSource = &Source[string]{}

type SyncSource interface {
	ReadRupture() *Rupture
	Current() any
	CurrentKey() string
	NextKey() string
	IsSynchronized() bool
	ComputeSynchro(string)
	Read() error
}

type Source[T any] struct {
	next    T
	current T

	keyer  func(T) string
	reader func() (T, error)

	readRupture *Rupture

	currentKey string
	nextKey    string

	synchronized bool
}

func NewSource[T any](reader func() (T, error), keyer func(T) string, readRupture *Rupture) *Source[T] {
	return &Source[T]{
		synchronized: true,
		keyer:        keyer,
		reader:       reader,
		readRupture:  readRupture,
	}
}

func (s *Source[T]) ReadRupture() *Rupture {
	return s.readRupture
}

func (s *Source[T]) Current() any {
	return s.current
}

func (s *Source[T]) CurrentKey() string {
	return s.currentKey
}

func (s *Source[T]) NextKey() string {
	return s.nextKey
}

func (s *Source[T]) IsSynchronized() bool {
	return s.synchronized
}

func (s *Source[T]) Read() error {
	if !s.synchronized {
		return nil
	}

	if s.readRupture != nil && !s.readRupture.last {
		return nil
	}

	return s.read()
}

func (s *Source[T]) ComputeSynchro(key string) {
	compareKey := s.currentKey[:min(len(key), len(s.currentKey))]
	s.synchronized = compareKey == key[:len(compareKey)]
}

func (s *Source[T]) read() error {
	s.current = s.next
	s.currentKey = s.nextKey

	next, err := s.reader()
	if err != nil && err != io.EOF {
		return err
	}

	s.next = next
	if err == io.EOF {
		s.nextKey = finalValue
	} else {
		s.nextKey = s.keyer(s.next)
	}

	return nil
}

func NewSliceSource[T any](arr []T, keyer func(T) string, readRupture *Rupture) *Source[T] {
	index := -1

	return NewSource(func() (output T, err error) {
		index++
		if index < len(arr) {
			output = arr[index]
		} else {
			err = io.EOF
		}

		return
	}, keyer, readRupture)
}

func NewChanSource[T any](input <-chan T, keyer func(T) string, readRupture *Rupture) *Source[T] {
	var ok bool

	return NewSource(func() (output T, err error) {
		if output, ok = <-input; !ok {
			err = io.EOF
		}

		return
	}, keyer, readRupture)
}

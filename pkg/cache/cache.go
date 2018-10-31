package cache

import (
	"sync"
	"time"
)

// TimeMap allows to store content in a map with entry duration life
type TimeMap interface {
	Store(key, value interface{}, duration time.Duration)
	Load(key interface{}) (interface{}, bool)
	LoadOrStore(key, value interface{}, duration time.Duration) (interface{}, bool)
	Range(f func(key, value interface{}) bool)
	Delete(key interface{})
}

// New create a new TimeMap
func New() TimeMap {
	return &timeMap{
		store: sync.Map{},
	}
}

type mapValue struct {
	content    interface{}
	expiration time.Time
}

func (t *mapValue) isValid() bool {
	return t.expiration.After(time.Now())
}

var _ TimeMap = &timeMap{}

type timeMap struct {
	store sync.Map
}

// Store content to map with given duration
func (t *timeMap) Store(key, value interface{}, duration time.Duration) {
	t.store.Store(key, mapValue{
		content:    value,
		expiration: time.Now().Add(duration),
	})
}

// Load content from map
func (t *timeMap) Load(key interface{}) (interface{}, bool) {
	value, ok := t.store.Load(key)

	if ok {
		timeValue := value.(mapValue)
		if timeValue.isValid() {
			return timeValue.content, true
		}

		t.Delete(key)
	}

	return nil, false
}

// LoadOrStore given content to the map
func (t *timeMap) LoadOrStore(key, value interface{}, duration time.Duration) (interface{}, bool) {
	if content, ok := t.Load(key); ok {
		return content, false
	}

	t.Store(key, value, duration)
	return value, true
}

// Range browser all entries of map
func (t *timeMap) Range(f func(key, value interface{}) bool) {
	t.store.Range(func(key, value interface{}) bool {
		timeValue := value.(mapValue)
		if timeValue.isValid() {
			return f(key, value)
		}

		return true
	})
}

// Delete content from map
func (t *timeMap) Delete(key interface{}) {
	t.store.Delete(key)
}

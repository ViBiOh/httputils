package breaksync

import "fmt"

func ruptureExtractSimple(a string) string {
	return a
}

func sourceBasicKeyer(e interface{}) string {
	return fmt.Sprintf("%#v", e)
}

func newSliceSource(slice []interface{}, keyer func(interface{}) string, readRupture *Rupture) *Source {
	index := -1

	return NewSource(func() (interface{}, error) {
		index++
		if index < len(slice) {
			return slice[index], nil
		}
		return nil, nil
	}, keyer, readRupture)
}

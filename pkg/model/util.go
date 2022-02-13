package model

import "reflect"

// IsNil check if an interface is nil or not
func IsNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

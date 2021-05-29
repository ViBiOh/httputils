package breaksync

import "fmt"

func ruptureExtractSimple(a string) string {
	return a
}

func sourceBasicKeyer(e interface{}) string {
	return fmt.Sprintf("%#v", e)
}

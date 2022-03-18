package breaksync

import "fmt"

func ruptureExtractSimple(a string) string {
	return a
}

func sourceBasicKeyer(e any) string {
	return fmt.Sprintf("%#v", e)
}

package tools

import "unicode"

// ToCamel change first letter to lowerCase
func ToCamel(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

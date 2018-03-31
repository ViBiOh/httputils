package tools

import "unicode"

// ToCamel change first letter to lowerCase
func ToCamel(s string) string {
	if len(s) == 0 {
		return s
	}

	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

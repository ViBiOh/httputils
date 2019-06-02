package tools

import (
	"strings"
	"unicode"
)

// ToCamel change first letter to lowerCase
func ToCamel(s string) string {
	if len(s) == 0 {
		return s
	}

	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

// IncludesString checks in an array includes given string
func IncludesString(array []string, lookup string) bool {
	for _, item := range array {
		if strings.EqualFold(item, lookup) {
			return true
		}
	}

	return false
}

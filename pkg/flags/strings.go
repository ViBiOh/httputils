package flags

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	upperCaseRegex = regexp.MustCompile(`(?m)([A-Z])([A-Z]*)`)
)

func changeFirstCase(s string, upper bool) string {
	if len(s) == 0 {
		return s
	}

	a := []rune(s)
	if upper {
		a[0] = unicode.ToUpper(a[0])
	} else {
		a[0] = unicode.ToLower(a[0])
	}

	return string(a)
}

// FirstUpperCase change first letter to UpperCase
func FirstUpperCase(s string) string {
	return changeFirstCase(s, true)
}

// FirstLowerCase change first letter to lowerCase
func FirstLowerCase(s string) string {
	return changeFirstCase(s, false)
}

// SnakeCase transform camelCase to snake_case
func SnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	snaked := upperCaseRegex.ReplaceAllString(s, "_$1$2")
	if snaked[0] == '_' {
		return snaked[1:]
	}

	return strings.ReplaceAll(snaked, "-", "_")
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

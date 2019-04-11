package db

import (
	"fmt"
	"strings"
)

// PrepareFullTextSearch replace $INDEX param in query and expand words
func PrepareFullTextSearch(query, search string, index uint) (string, string) {
	if search == "" {
		return "", ""
	}

	words := strings.Split(search, " ")
	transformedWords := make([]string, 0, len(words))

	for _, word := range words {
		transformedWords = append(transformedWords, word+":*")
	}

	return strings.Replace(query, "$INDEX", fmt.Sprintf("$%d", index), -1), strings.Join(transformedWords, " | ")
}

package usecase

import (
	"strings"
	"unicode"
)

func Tokenize(s string) []string {
	var words []string
	var cur strings.Builder

	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(unicode.ToLower(r))
		} else {
			flush()
		}
	}
	flush()

	return words
}

package util

import (
	"strings"
	"unicode"
)

func Exist[T ~int | ~int32 | ~uint32 | ~int64 | ~string](arr []T, n T) bool {
	for _, i := range arr {
		if i == n {
			return true
		}
	}
	return false
}

// IsAlphanumeric special character check
func IsAlphanumeric(input string) bool {
	for _, r := range input {
		if !(unicode.Is(unicode.Han, r) || // Chinese character
			unicode.IsLetter(r) || // English letter
			unicode.IsDigit(r) || // Digit
			r == ' ') { // Space
			if r == ':' || r == '"' || r == '\'' { // Check for specific unwanted characters
				return false
			}
		}
		if unicode.IsUpper(r) {
			return false
		}
	}
	return !strings.ContainsAny(input, "~#@$%^&*()<>,.{}[]、|/？?;'!！=+")
}

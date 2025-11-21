package util

import (
	"math"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func UUID() string {
	return uuid.NewV4().String()
}

func Reverse[T any](lst []T) {
	// reverse
	for i, j := 0, len(lst)-1; i < j; i, j = i+1, j-1 {
		lst[i], lst[j] = lst[j], lst[i]
	}
}

func InLst(item string, lst []string) bool {
	item = strings.TrimSpace(item)
	for _, l := range lst {
		if item == strings.TrimSpace(l) {
			return true
		}
	}

	return false
}

// Int type division. The first return value is the quotient, the second is the remainder
func DivideAndRemainder(dividend, divisor int) (quotient, remainder int) {
	quotient = int(math.Floor(float64(dividend) / float64(divisor)))
	remainder = dividend - quotient*divisor
	return quotient, remainder
}

// len has a total of 101 sets of data, 8 per page, so 12+1=13 pages are needed
func GetPageTotal(len, pageSize int) int {
	q, r := DivideAndRemainder(len, pageSize)
	if r == 0 {
		return q
	}
	return q + 1
}

// The Intersection function is used to find the intersection of two slices of comparable types.
func Intersection[T comparable](slice1, slice2 []T) []T {
	m := make(map[T]struct{}) // Use empty structures to save space
	var result []T

	// Put all elements of the first slice into map
	for _, item := range slice1 {
		m[item] = struct{}{}
	}

	// Check if the element in the second slice exists in the map
	for _, item := range slice2 {
		if _, ok := m[item]; ok {
			// If it exists and the element does not exist in the result set, add it to the result set.
			if !contains(result, item) {
				result = append(result, item)
			}
		}
	}

	return result
}

// The contains function checks whether a slice contains an element.
func contains[T comparable](slice []T, elem T) bool {
	for _, item := range slice {
		if item == elem {
			return true
		}
	}
	return false
}

package worker

import (
	"golang.org/x/exp/constraints"
	"sort"
)

// GetRecoverError extracts an error from a recoverable panic.
// It checks if the recovered value is an error type, and if so, returns it.
// If the recovered value is not an error type, it returns nil.
func GetRecoverError(rec any) error {
	// Check if recoverable value is not nil
	if rec != nil {
		// Type switch on the recovered value
		switch e := rec.(type) {
		// If recovered value is of type error
		case error:
			return e

		// If recovered value is of any other type
		default:
			return nil
		}
	} else {
		// If recoverable value is nil
		return nil
	}
}

// Contains checks if the provided element is present in the slice.
// It first sorts the slice and then performs a binary search to determine if the element exists.
// Returns true if the element is found, otherwise false.
func Contains[T constraints.Ordered](elements []T, element T) bool {
	// Check if the slice is nil. If it is, return false because there's nothing to search.
	if elements == nil {
		return false
	}

	// Sort the slice in ascending order.
	// Sorting is necessary for binary search to work correctly.
	sort.Slice(elements, func(i, j int) bool {
		return elements[i] < elements[j]
	})

	// Use binary search to find the index of the element.
	// `sort.Search` will return the index of the first element greater than or equal to `element`.
	// If no such element is found, it returns the length of the slice.
	index := sort.Search(len(elements), func(i int) bool {
		return elements[i] >= element
	})

	// Validate the index to ensure it's within the bounds of the slice.
	// Check if the element at the found index matches the search element.
	// Return true if the element at the index equals the search element, otherwise false.
	return index < len(elements) && elements[index] == element
}

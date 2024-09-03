package worker

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

// Exclude removes all instances of a specified value from the provided slice.
// It creates a new slice containing only the elements that are not equal to the specified value.
// This approach efficiently constructs the result slice by reusing the original slice's underlying array,
// avoiding unnecessary memory allocations.
func Exclude[T comparable](elements []T, element T) []T {
	// Initialize the result slice with the same underlying array as the original slice.
	// This avoids unnecessary allocations and keeps the capacity the same.
	result := elements[:0]

	// Iterate over each item in the original slice.
	for _, item := range elements {
		// Check if the current item is not equal to the specified value to be excluded.
		if item != element {
			// Append the item to the result slice if it is not equal to the specified value.
			result = append(result, item)
		}
	}

	// Return the filtered slice with the specified value removed.
	return result
}

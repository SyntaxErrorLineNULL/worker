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

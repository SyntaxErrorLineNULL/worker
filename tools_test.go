package worker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetRecoverError tests the GetRecoverError function to ensure it correctly extracts errors from recoverable panics.
func TestGetRecoverError(t *testing.T) {
	t.Parallel()

	// RecoverableError tests the behavior of the GetRecoverError function when it receives a specific error as input.
	// It verifies that the function correctly returns the same error without modification. This ensures that
	// GetRecoverError is functioning as expected in scenarios where the input is a recoverable or known error.
	// By passing a sample error to the function and comparing the result with the original error, we confirm
	// that GetRecoverError handles the error consistently and accurately.
	t.Run("RecoverableError", func(t *testing.T) {
		// Create a new error with a sample message for testing.
		// This represents an error that will be passed to the GetRecoverError function.
		err := errors.New("sample error")

		// Call the GetRecoverError function with the created error.
		// This function is expected to process or handle the error in some way.
		result := GetRecoverError(err)

		// Assert that the result of GetRecoverError is equal to the original error.
		// This checks that GetRecoverError correctly returns the same error it received.
		assert.Equal(t, err, result, "Expected GetRecoverError to return the same error")
	})

	// RecoverableNonError tests the behavior of the GetRecoverError function when it receives a non-error value as input.
	// This test verifies that the function correctly returns nil when given a value that is not an error.
	// It ensures that GetRecoverError handles non-error inputs appropriately by not returning an unexpected error value.
	t.Run("RecoverableNonError", func(t *testing.T) {
		// Define a sample non-error value to be used in the test.
		// This represents a value that is not an error, to check how GetRecoverError processes such inputs.
		nonError := "sample non-error"

		// Call the GetRecoverError function with the non-error value.
		// The function is expected to handle non-error inputs and return nil in this case.
		result := GetRecoverError(nonError)

		// Assert that the result of GetRecoverError is nil.
		// This confirms that GetRecoverError properly returns nil for non-error inputs, as expected.
		assert.Nil(t, result, "Expected GetRecoverError to return nil for non-error recoverable value")
	})

	// NilRecoverable tests the behavior of the GetRecoverError function when it receives a nil value as input.
	// This test verifies that the function returns nil when given a nil recoverable value.
	// It ensures that GetRecoverError handles nil inputs appropriately and does not produce unexpected results.
	t.Run("NilRecoverable", func(t *testing.T) {
		// Call the GetRecoverError function with a nil value.
		// This represents a scenario where no error or recoverable value is provided.
		result := GetRecoverError(nil)

		// Assert that the result of GetRecoverError is nil.
		// This confirms that the function properly handles a nil input by returning nil, as expected.
		assert.Nil(t, result, "Expected GetRecoverError to return nil for nil recoverable value")
	})
}

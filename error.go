package worker

import "errors"

var (
	// WaitGroupIsNil is an error indicating that a nil wait group was provided.
	// This error is used to signal that an operation requiring a non-nil wait group
	// encountered an invalid input, preventing proper synchronization.
	WaitGroupIsNil = errors.New("wait group is nil")

	// ContextIsNil is an error indicating that a nil context was provided.
	// This error is used to signal that an operation requiring a non-nil context
	// encountered an invalid input, which could disrupt the operation's control flow.
	ContextIsNil = errors.New("context is nil")
)

// Error encapsulates an error along with the reference to the Worker instance
// where the error occurred. It is used to convey error information from a Worker
// to external components through the error channel.
type Error struct {
	Error    error  // The actual error that occurred.
	Instance Worker // Reference to the Worker instance where the error occurred.
}

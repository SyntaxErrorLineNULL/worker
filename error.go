package worker

import "errors"

var (
	// WaitGroupIsNilError is an error indicating that a nil wait group was provided.
	// This error is used to signal that an operation requiring a non-nil wait group
	// encountered an invalid input, preventing proper synchronization.
	WaitGroupIsNilError = errors.New("wait group is nil")

	// ContextIsNilError is an error indicating that a nil context was provided.
	// This error is used to signal that an operation requiring a non-nil context
	// encountered an invalid input, which could disrupt the operation's control flow.
	ContextIsNilError = errors.New("context is nil")

	// ChanIsCloseError is an error indicating that a channel has been closed.
	// This error is typically used when an operation attempts to send or receive
	// on a closed channel, which is an illegal operation in Go.
	ChanIsCloseError = errors.New("chan is close")

	// ChanIsEmptyError is an error indicating that a channel is empty.
	// This error might be used when an operation expected a channel to have available data,
	// but found it empty, preventing the operation from proceeding as intended.
	ChanIsEmptyError = errors.New("chan is empty")

	// MaxWorkersReachedError is an error indicating that the maximum number of workers has been reached.
	// This error is used to prevent the creation of additional workers beyond the allowed limit,
	// ensuring that the system's resources are managed efficiently and within expected constraints.
	MaxWorkersReachedError = errors.New("maximum number of workers reached")
)

// Error encapsulates an error along with the reference to the Worker instance
// where the error occurred. It is used to convey error information from a Worker
// to external components through the error channel.
type Error struct {
	Error    error  // The actual error that occurred.
	Instance Worker // Reference to the Worker instance where the error occurred.
}

package worker

import "errors"

var (
	// ErrContextIsNil is an error indicating that a nil context was provided.
	// This error is used to signal that an operation requiring a non-nil context
	// encountered an invalid input, which could disrupt the operation's control flow.
	ErrContextIsNil = errors.New("context is nil")

	// ErrChanIsClose is an error indicating that a channel has been closed.
	// This error is typically used when an operation attempts to send or receive
	// on a closed channel, which is an illegal operation in Go.
	ErrChanIsClose = errors.New("chan is close")

	// ErrChanIsEmpty is an error indicating that a channel is empty.
	// This error might be used when an operation expected a channel to have available data,
	// but found it empty, preventing the operation from proceeding as intended.
	ErrChanIsEmpty = errors.New("chan is empty")

	// ErrMaxWorkersReached is an error indicating that the maximum number of workers has been reached.
	// This error is used to prevent the creation of additional workers beyond the allowed limit,
	// ensuring that the system's resources are managed efficiently and within expected constraints.
	ErrMaxWorkersReached = errors.New("maximum number of workers reached")

	// ErrWorkerIsNil is an error that indicates an operation was attempted with a nil worker.
	// This error is used to signal that a worker instance expected to be valid was found to be nil,
	// and therefore the operation cannot proceed.
	ErrWorkerIsNil = errors.New("worker is nil")

	// ErrWorkerPoolStop is an error that indicates an operation was attempted on a stopped worker pool.
	// This error is used to signal that the worker pool has been stopped, meaning no further work can be assigned
	// or processed, and the operation should be terminated.
	ErrWorkerPoolStop = errors.New("worker pool is stopped")
)

// Error encapsulates an error along with the reference to the Worker instance
// where the error occurred. It is used to convey error information from a Worker
// to external components through the error channel.
type Error struct {
	Error    error  // The actual error that occurred.
	Instance Worker // Reference to the Worker instance where the error occurred.
}

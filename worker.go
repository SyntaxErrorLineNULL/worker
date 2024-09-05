package worker

import (
	"context"
	"sync"
)

// Worker defines the interface for a worker in a worker pool system.
// A Worker is responsible for processing tasks from a queue, managing its
// lifecycle (start, stop), and providing status and error information.
//
//go:generate mockery --name=Worker
type Worker interface {
	// String returns a string representation of the Worker.
	// Typically, this is the worker's name, which is useful for logging and debugging purposes.
	String() string

	// SetContext assigns a context to the worker.
	// The context is used to control the worker's execution and can be used to
	// cancel operations or signal timeouts. Returns an error if the context is nil
	// or if there's an issue setting the context.
	//
	// NOTE: Context is passed from the worker pool (parent) to manage the state of the worker
	SetContext(ctx context.Context) error

	// SetQueue assigns a task queue to the worker.
	// The worker will listen to this queue for incoming tasks to process.
	// Returns an error if the queue is nil or invalid.
	//
	// NOTE: The channel must be closed after the worker pool is stopped.
	// Otherwise all the workers will simply be stopped and stop working.
	SetQueue(queue chan Task) error

	// Restart attempts to restart the worker by incrementing the retry count
	// and then invoking the Start method to resume the worker's operation.
	// This method is used to recover a worker that may have encountered an issue,
	// tracking the number of recovery attempts.
	Restart(wg *sync.WaitGroup)

	// Start begins the worker's operation, processing tasks from the assigned queue.
	// It requires a sync.WaitGroup to manage the concurrent execution of workers.
	// The WaitGroup is used to ensure that all workers complete their tasks before
	// the program exits or the pool is shut down.
	Start(wg *sync.WaitGroup)

	// Stop initiates the worker's shutdown process.
	// This method returns a channel that is closed when the worker has completely
	// stopped. It allows for graceful shutdowns by signaling when it is safe to
	// release resources or proceed with other operations.
	Stop() <-chan struct{}

	// GetStatus retrieves the current status of the worker.
	// The status provides information about the worker's state, such as whether
	// it is running, idle, or stopped. This is useful for monitoring and managing
	// the worker's activity.
	GetStatus() Status

	// GetError returns a channel that the worker uses to report errors.
	// The channel carries errors encountered during task processing. This allows
	// the system to log, handle, or react to errors in a centralized manner.
	GetError() chan *Error

	// GetRetry returns the current retry count for the worker.
	// The retry count indicates the number of attempts made to restart the worker
	// in an effort to restore its operation after encountering an issue.
	GetRetry() int32
}

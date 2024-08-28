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
	// SetContext assigns a context to the worker.
	// The context is used to control the worker's execution and can be used to
	// cancel operations or signal timeouts. Returns an error if the context is nil
	// or if there's an issue setting the context.
	SetContext(ctx context.Context) error

	// SetQueue assigns a task queue to the worker.
	// The worker will listen to this queue for incoming tasks to process.
	// Returns an error if the queue is nil or invalid.
	SetQueue(queue chan Task) error

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
}

package worker

import (
	"context"
	"sync"
)

// Task defines the contract for a task that can be executed, monitored, and managed.
// It provides methods for setting context, wait groups, and completion channels, as well as
// methods for handling errors and stopping execution. Implementations of this interface are
// expected to provide specific behavior for processing tasks, handling interruptions, and
// managing execution state.
//
//go:generate mockery --name=Task
type Task interface {
	// SetWaitGroup assigns a sync.WaitGroup to the task.
	// This allows the task to signal completion when it has finished its execution.
	// The wait group is used to synchronize with other goroutines, ensuring that all tasks
	// complete before proceeding.
	SetWaitGroup(wg *sync.WaitGroup) error

	// SetDoneChannel sets the channel that will be used to signal when the task is done.
	// This channel is expected to be closed once the task completes its execution. It is
	// used to notify other parts of the system that the task has finished.
	SetDoneChannel(done chan struct{}) error

	// SetContext assigns a context to the task.
	// The context can be used to manage task execution, including handling cancellation and
	// timeouts. It allows the task to respond to external signals for stopping or modifying
	// its behavior.
	SetContext(ctx context.Context) error

	// GetError retrieves the error encountered during task execution, if any.
	// This method allows checking for errors that occurred during the task's run and helps
	// in debugging and error handling.
	GetError() error

	// String returns a string representation of the task.
	// This is useful for logging and debugging purposes, providing a textual description
	// of the task to aid in understanding its state and behavior during execution.
	String() string

	// Run starts the execution of the task.
	// This method should contain the logic for performing the task's work. It is typically
	// called in a separate goroutine to allow asynchronous execution of the task.
	Run()

	// Stop signals the task to stop executing.
	// This method is used to gracefully terminate the task before it completes. It should
	// handle cleanup and termination logic to ensure the task is stopped in a controlled manner.
	Stop()
}

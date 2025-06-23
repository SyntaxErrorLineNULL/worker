package worker

import (
	"context"
	"sync"
	"time"
)

// Task defines the contract for a task that can be executed, monitored, and managed.
// It provides methods for setting context, wait groups, and completion channels, as well as
// methods for handling errors and stopping execution. Implementations of this interface are
// expected to provide specific behavior for processing job, handling interruptions, and
// managing execution state.
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
	GetError() <-chan error

	// String returns a string representation of the task.
	// This is useful for logging and debugging purposes, providing a textual description
	// of the task to aid in understanding its state and behavior during execution.
	String() string

	// Run starts the execution of the task.
	// This method should contain the logic for performing the task's work. It is typically
	// called in a separate goroutine to allow asynchronous execution of the task.
	Run(workerTimeout time.Duration)

	// Stop signals the task to stop executing.
	// This method is used to gracefully terminate the task before it completes. It should
	// handle cleanup and termination logic to ensure the task is stopped in a controlled manner.
	Stop()
}

// Pool defines the interface for managing a pool of worker goroutines.
// It provides methods to start the pool, manage tasks, control workers, and handle shutdowns.
// A Pool is responsible for initializing and managing a set of worker goroutines that process jobs concurrently.
// It includes methods for adding task and workers, retrieving the count of running workers, and stopping the entire pool.
// This interface facilitates efficient task processing and dynamic scaling of worker resources.
type Pool interface {
	// Run starts the worker pool and initializes worker goroutines to process jobs.
	// This method sets up the worker pool, starts all the worker goroutines, and prepares them to begin processing jobs.
	// It ensures that the pool is in an operational state and ready to handle incoming tasks.
	Run()

	// AddTaskInQueue attempts to add a new task to the pool's task queue for processing by the workers.
	// If the task cannot be added due to the pool being closed or other reasons, an error is returned.
	// This method ensures that the task is handled in a non-blocking way and leverages context-based
	// cancellation or timeout handling if needed.
	AddTaskInQueue(task Task) error

	// AddWorker adds a new worker to the worker pool.
	// It creates a new worker instance, increments the worker count, and starts the worker in a separate goroutine.
	// The worker is added to the pool's internal workers slice, and the worker count is incremented using the pool's mutex.
	// The wait group (wg) is also incremented to track the new worker's lifecycle.
	// Returns an error if the worker count cannot be incremented.
	AddWorker(worker Worker) error

	// RunningWorkers returns the number of currently active worker goroutines in the pool.
	// This method provides a count of workers that are actively processing jobs and have not been stopped.
	// It is useful for monitoring the state of the worker pool and ensuring that the desired number of workers are running.
	// Returns:
	// - int: The number of active workers.
	RunningWorkers() int32

	// Stop terminates the worker pool and all its associated workers.
	// It ensures a clean shutdown by canceling the pool's context, signaling workers to stop,
	// and waiting for all workers to finish before completing the shutdown sequence.
	Stop()
}

// Worker represents an interface for a worker in a worker pool system.
// A worker is responsible for processing jobs assigned to it and managing its own lifecycle.
// It allows external components to start, stop, and monitor the worker's status and errors.
type Worker interface {
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
	// Otherwise, all the workers will simply be stopped and stop working.
	SetQueue(queue chan Task) error

	// SetWorkerErrChannel assigns an error channel to the worker.
	// This channel is used to report serious errors or panics that occur
	// during task processing. The pool listens to this channel to detect
	// when a worker encounters a severe issue and needs to be restarted.
	// Returns an error if the channel is closed at the time of assignment.
	SetWorkerErrChannel(errCh chan *Error) error

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

	// IsStop checks if the worker is currently stopped.
	// This method allows external components to determine the worker's state,
	// ensuring accurate monitoring and control over its lifecycle.
	IsStop() bool

	// Stop initiates the worker's shutdown process.
	// This method returns a channel that is closed when the worker has completely
	// stopped. It allows for graceful shutdowns by signaling when it is safe to
	// release resources or proceed with other operations.
	Stop() <-chan struct{}

	// GetError returns a channel that the worker uses to report errors.
	// The channel carries errors encountered during task processing. This allows
	// the system to log, handle, or react to errors in a centralized manner.
	GetError() chan *Error

	// GetRetry returns the current retry count for the worker.
	// The retry count indicates the number of attempts made to restart the worker
	// in an effort to restore its operation after encountering an issue.
	GetRetry() int32
}

package worker

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SyntaxErrorLineNULL/worker"
)

// Worker represents a worker in a worker pool. It performs tasks assigned to it
// by the pool's dispatcher. The worker is designed to run concurrently with other
// workers, processing tasks in parallel.
type Worker struct {
	workerName     string             // Unique identifier for the worker.
	workerContext  context.Context    // Context for the worker's operations.
	mutex          sync.RWMutex       // Mutex to control access to shared resources.
	stopCh         chan struct{}      // Channel to signal the worker to stop processing jobs.
	queue          <-chan worker.Task // Channel to receive jobs from the pool's dispatcher.
	currentProcess worker.Task        // Current job being processed by the worker.
	timeout        time.Duration      // Timeout duration for processing a job.
	errCh          chan *worker.Error // Channel to send and receive errors that occur in the worker.
	onceStop       sync.Once          // Ensures the stop currentProcess is only executed once.
	retryCount     atomic.Int32       // Number of attempts to bring the worker back to life.
	isStop         atomic.Bool
	logger         *log.Logger
}

// NewWorker initializes a new Worker instance with the provided workerName.
// It sets up necessary channels and a logger for the worker, and returns a pointer to the Worker.
func NewWorker(workerName string, timeout time.Duration, logger *log.Logger) *Worker {
	// return new worker instance.
	return &Worker{workerName: workerName, timeout: timeout, logger: logger, stopCh: make(chan struct{}, 1)}
}

// String returns the name of the worker as its string representation.
// This method allows the Worker instance to be represented as a string, which is useful
// for logging, debugging, or any other context where the worker's name is needed in text format.
func (w *Worker) String() string {
	// Return the worker's name, which is stored in the workerName field.
	// This ensures that the String method provides a meaningful and identifiable
	// representation of the Worker instance.
	return w.workerName
}

// SetContext sets the context for the worker. This method is used to provide
// a new context for the worker, which can be used to control its operations
// and manage its lifecycle. The method ensures that the provided context is not
// nil before setting it, maintaining the integrity of the worker's context.
func (w *Worker) SetContext(ctx context.Context) error {
	// Check if the provided context is nil. A nil context is invalid and
	// should not be used. Return an error in this case to prevent setting
	// an invalid context for the worker.
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	// Assign the provided context to the worker's context field.
	// This allows the worker to use this context for its operations.
	w.workerContext = ctx

	return nil
}

// SetQueue sets the job queue channel for the worker. This method allows
// the worker to be assigned a new job queue channel, which it will use
// to receive jobs. The method ensures that the provided channel is open before
// setting it, returning an error if the channel is closed.
func (w *Worker) SetQueue(queue chan worker.Task) error {
	// Use a select statement with a default case to check if the provided channel is closed.
	// The select statement attempts to receive from the queue channel.
	select {
	// Attempt to receive from the queue channel.
	case _, ok := <-queue:
		// If the receive operation fails, the channel is closed.
		// Return an error indicating that the channel is closed.
		if !ok {
			return errors.New("queue cannot be close")
		}
	// If the receive operation would block, continue without doing anything.
	default:
	}

	// Set the worker's queue channel to the provided queue channel.
	w.queue = queue

	// Return nil indicating that the operation was successful.
	return nil
}

// SetWorkerErrChannel assigns the provided error channel to the worker for reporting
// serious panic errors that occur during its operation. This channel is used to notify
// the worker pool of such errors, so the pool can take appropriate action, such as
// restarting the worker.
func (w *Worker) SetWorkerErrChannel(errCh chan *worker.Error) error {
	// Use a select statement with a default case to check if the provided channel is closed.
	// The select statement attempts to receive from the queue channel.
	select {
	// Attempt to receive from the queue channel.
	case _, ok := <-errCh:
		// If the receive operation fails, the channel is closed.
		// Return an error indicating that the channel is closed.
		if !ok {
			return errors.New("error channel cannot be nil")
		}
	// If the receive operation would block, continue without doing anything.
	default:
	}

	// Set the error channel to the provided queue channel.
	w.errCh = errCh

	// Return nil indicating that the operation was successful.
	return nil
}

// Restart attempts to restart the worker by incrementing the retry count
// and then invoking the Start method to resume the worker's operation.
// The retry count is incremented to track the number of recovery attempts.
func (w *Worker) Restart(wg *sync.WaitGroup) {
	w.logger.Printf("restart worker: %s", w.workerName)
	// Increment the retry count to indicate a new recovery attempt.
	w.retryCount.Add(1)

	// Start the worker again.
	go w.Start(wg)
}

// Start method is a goroutine that continuously extracts and processes tasks assigned to a worker, ensuring it's always ready to handle incoming tasks.
// We create a specific logger for the worker to track its activities.
// The method operates in a loop, constantly handling various scenarios to exit gracefully, such as receiving stop signals or context termination.
// When a task is available, the worker retrieves and logs it.
// The worker sets its status to StatusWorkerRunning, indicating active task processing.
// It executes the task using the Run method, passing worker-specific context.
// Any task execution errors are logged for debugging.
// After task completion, the worker resets its status to StatusWorkerIdle, indicating readiness for more tasks.
// A log message signals the task processing completion.
// This method guarantees continuous task processing while maintaining detailed logs for debugging and monitoring.
func (w *Worker) Start(wg *sync.WaitGroup) {
	// This deferred function serves as a recovery and cleanup mechanism for the worker.
	// If a panic occurs during the execution of a worker's main cycle,
	// in order not to lose a worker, the error channel will receive information with the error and the instance of the worker to be recovered.
	defer func() {
		// Defer a function to recover from any panic that occurs during shutdown.
		// If a panic is recovered, it attempts to log the error
		if rec := recover(); rec != nil {
			// Convert the recovered panic value into an error.
			// This ensures that any panic during task addition is properly handled.
			if err := worker.GetRecoverError(rec); err != nil {
				// Send the error to the worker's error channel for external handling.
				w.errCh <- &worker.Error{Error: err, Instance: w}
			}
		}

		// If a wait group is provided, decrement its counter to signal that the worker has completed its task.
		if wg != nil {
			// Decrement the WaitGroup counter to signal that the worker has completed its task.
			wg.Done()
		}
	}()

	// Start a goroutine to currentProcess jobs. The worker will continuously listen for new jobs,
	// stop signals, or context cancellation in this loop.
	for {
		// The select statement waits for one of its cases to be ready to execute.
		select {
		// This case handles the situation where the worker receives a signal from its stop channel.
		// The stop channel stopCh is used to signal that the worker should stop its execution.
		case <-w.stopCh:
			w.logger.Printf("stop channel, worker: %s", w.workerName)
			// Exit the loop, effectively stopping the worker's execution.
			// This happens when the stop channel is triggered, signaling that the worker should terminate.
			return

		// This case handles the situation where the worker's parent context is done.
		// The worker listens to the `workerContext` channel for a done signal,
		// which indicates that the context in which the worker operates has been cancelled or expired.
		case <-w.workerContext.Done():
			w.logger.Print("parent context is close")
			// Set the worker status to stopped.
			// This action updates the worker's status to reflect that it is stopping due to the parent context being done.
			// This helps in maintaining accurate status information and allows other components
			// to be aware that the worker is no longer active.
			w.Stop()

			// Exit the loop, effectively stopping the worker's execution.
			// The `return` statement breaks out of the infinite loop and stops further processing.
			// This ensures that the worker ceases its operations when the parent context is cancelled,
			// allowing it to exit gracefully and freeing up resources.
			return

		// This case handles incoming tasks from the worker's task queue.
		// The <-w.queue operation attempts to receive a task from the channel.
		// The ok variable indicates whether the channel is still open true or has been closed false.
		case task, ok := <-w.queue:
			// Check if the job channel is closed.
			if ok {
				// Check if the received task is not nil.
				if task != nil {
					// Assign the received task to the worker's `currentProcess` field.
					// This keeps track of the task currently being executed by the worker.
					w.currentProcess = task

					// Increment the WaitGroup counter to account for the new task being processed.
					// This helps synchronize the completion of the task with other concurrent operations.
					wg.Add(1)

					// Set the WaitGroup for the task. This allows the task to signal when it has completed.
					// The `task.SetWaitGroup(workerWg)` call ensures that the task can signal its completion.
					_ = task.SetWaitGroup(wg)

					// Execute the task's `Run` method.
					// This method contains the logic to currentProcess the task.
					task.Run(w.timeout)

					// After the task completes, reset the worker's `currentProcess` to `nil`.
					// This clears the reference to the completed task.
					w.currentProcess = nil
				}
			} else {
				// If the task queue channel is closed
				w.logger.Printf("task queue is close: worker: %s", w.workerName)
				// Call the worker's `Stop` method to clean up and stop the worker.
				// This method sets the worker status to stopped and performs necessary cleanup.
				w.Stop()

				// Exit the loop and terminate the worker's execution.
				// This `return` statement breaks out of the loop, effectively stopping the worker.
				return
			}
		}
	}
}

// Stop signals the worker to stop processing tasks and returns a channel to indicate completion.
// It closes the stop channel, causing the worker to exit its processing loop and finish the current job.
// If there is a task in processing at the time of worker termination, it will be stopped.
func (w *Worker) Stop() <-chan struct{} {
	defer func() {
		// Defer a function to recover from any panic that occurs during shutdown.
		// If a panic is recovered, it attempts to log the error
		if rec := recover(); rec != nil {
			// Convert the recovered panic value into an error.
			// This ensures that any panic during task addition is properly handled.
			if err := worker.GetRecoverError(rec); err != nil {
				// This is in case something caused a panic, but the worker status was not set to true,
				// so that the Worker pool would not recover the worker.
				if !w.IsStop() {
					w.isStop.Store(true)
				}
			}
		}
	}()

	// Create a channel to signal when the worker has stopped.
	// The buffered channel allows sending a single signal indicating the stop currentProcess is complete.
	doneCh := make(chan struct{}, 1)

	// Ensure the stop currentProcess is only executed once using sync.Once.
	w.onceStop.Do(func() {
		// Acquire a read lock on the worker's mutex to ensure thread-safe access during the shutdown sequence.
		// This lock prevents concurrent operations from interfering with the worker's state while it is being shut down.
		w.mutex.RLock()
		// Schedule the release of the read lock once the shutdown sequence is complete.
		// Using defer ensures that the lock is always released, even if an error or panic occurs during shutdown.
		defer w.mutex.RUnlock()

		// If there is a task currently being processed, stop it.
		// This ensures that any ongoing work is properly terminated.
		if w.currentProcess != nil {
			w.currentProcess.Stop()
		}

		// Reset the currentProcess to nil, indicating that the worker is no longer actively processing a task.
		// This step ensures that the worker is in a clean state after the shutdown and ready for potential reuse.
		w.currentProcess = nil

		// Send a signal through the doneCh channel to notify that the worker has stopped.
		// This signal can be used to inform other components that the worker has completed its shutdown.
		doneCh <- struct{}{}
		// Close the doneCh channel to indicate that no more signals will be sent,
		// signaling the completion of the worker's shutdown.
		close(doneCh)

		// Send a signal through the stopCh channel to notify that the worker should stop processing.
		// This ensures that the worker halts any additional operations.
		w.stopCh <- struct{}{}
		// Close the stopCh channel to indicate that no further stop signals will be sent,
		// marking the end of the worker's operational lifecycle.
		close(w.stopCh)
	})

	// Set the worker's stopped status to true to indicate that the worker is no longer active.
	// This ensures that external systems are aware that the worker has completed its tasks and is shut down.
	w.isStop.Store(true)

	// Return the done channel to provide a signal mechanism for external monitoring.
	// This allows other components to observe when the worker has fully completed its shutdown process.
	return doneCh
}

// IsStop checks if the worker has been stopped.
// This method provides a thread-safe way to determine the current state of the worker.
// The state is managed using an atomic operation, ensuring accurate checks even
// in highly concurrent environments. It returns true if the worker is stopped and
// false otherwise.
func (w *Worker) IsStop() bool {
	// Load the current value of the isStop atomic variable.
	// This operation is thread-safe, making it suitable for use in concurrent scenarios.
	return w.isStop.Load()
}

// GetError returns the channel through which worker errors are communicated.
// This allows external components to listen for and handle errors generated by the worker.
// The channel is used to send instances of Error, containing information about the error and the worker instance.
// Note:In some cases I was able to get panic when running `Job`. This is a very critical situation,
// I could not control all the workers so that in case of a panic I would not lose all the workers.
// Now workers in case of panic can notify the worker-pool that controls them and that worker-pool
// will restore a particular worker in case the pool is not stopped.
// This should help avoid problems, especially since we might lose all workers.
func (w *Worker) GetError() chan *worker.Error {
	// Return the error channel associated with the worker.
	return w.errCh
}

// GetRetry returns the current retry count for the worker.
// The retry count indicates the number of attempts made to restart the worker
// in an effort to restore its operation if it encountered an issue.
func (w *Worker) GetRetry() int32 {
	// Load and return the current value of retryCount.
	return w.retryCount.Load()
}

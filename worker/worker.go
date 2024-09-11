package worker

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/SyntaxErrorLineNULL/worker"
)

// Worker represents a worker in a task processing pool.
// It holds all necessary information and channels to process tasks, manage its state, and handle errors.
type Worker struct {
	workerName     string             // Unique identifier for the worker.
	workerContext  context.Context    // Context for the worker's operations.
	mutex          sync.RWMutex       // Mutex to control access to shared resources.
	stopCh         chan struct{}      // Channel to signal the worker to stop processing tasks.
	queue          <-chan worker.Task // Channel to receive jobs from the pool's dispatcher.
	currentProcess worker.Task        // Current task being processed by the worker.
	status         worker.Status      // Current status of the worker (e.g., running, stopped).
	errCh          chan *worker.Error // Channel to send and receive errors that occur in the worker.
	onceStop       sync.Once          // Ensures the stop process is only executed once.
	retryCount     atomic.Int32       // Number of attempts to bring the worker back to life.
	logger         *log.Logger
}

// NewWorker initializes a new Worker instance with the provided workerName.
// It sets up necessary channels and a logger for the worker, and returns a pointer to the Worker.
func NewWorker(workerName string) *Worker {
	logger := log.New(os.Stdout, "pool:", log.LstdFlags)
	return &Worker{
		workerName: workerName,
		stopCh:     make(chan struct{}, 1),
		logger:     logger,
	}
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
		return worker.ContextIsNilError
	}

	// Assign the provided context to the worker's context field.
	// This allows the worker to use this context for its operations.
	w.workerContext = ctx

	return nil
}

// SetQueue sets the task queue channel for the worker. This method allows
// the worker to be assigned a new task queue channel, which it will use
// to receive tasks. The method ensures that the provided channel is open before
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
			return worker.ChanIsCloseError
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
			return worker.ChanIsCloseError
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
	// Increment the retry count to indicate a new recovery attempt.
	w.retryCount.Add(1)

	// Start the worker again.
	w.Start(wg)
}

// Start begins the worker's execution cycle. It initializes the worker's status,
// manages tasks from the job queue, and handles errors and context cancellations.
// The method uses a WaitGroup to signal when the worker has finished its work and includes
// mechanisms for recovery from panics to ensure that the worker continues operating smoothly
// even if unexpected errors occur.
func (w *Worker) Start(wg *sync.WaitGroup) {
	// w.logger.Printf("worker start: %s\n", w.workerName)
	// As soon as a worker is created, it is necessarily in the status of StatusWorkerIdle.
	// This indicates that the worker is ready but currently not processing any jobs.
	w.setStatus(worker.StatusWorkerIdle)

	// This deferred function serves as a recovery and cleanup mechanism for the worker.
	// If a panic occurs during the execution of a worker's main cycle,
	// in order not to lose a worker, the error channel will receive information with the error and the instance of the worker to be recovered.
	defer func() {
		// Attempt to recover from a panic and retrieve the error.
		if rec := recover(); rec != nil {
			// Convert the recovered value to an error.
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

	// Infinite loop that allows the worker to continuously check for and process tasks.
	for {
		// The select statement waits for one of its cases to be ready to execute.
		select {
		// This case handles the situation where the worker receives a signal from its stop channel.
		// The stop channel stopCh is used to signal that the worker should stop its execution.
		case <-w.stopCh:
			w.logger.Printf("stop channel workerName: %s", w.workerName)
			// Exit the loop, effectively stopping the worker's execution.
			// This happens when the stop channel is triggered, signaling that the worker should terminate.
			return

		// This case handles the situation where the worker's parent context is done.
		// The worker listens to the `workerContext` channel for a done signal,
		// which indicates that the context in which the worker operates has been cancelled or expired.
		case <-w.workerContext.Done():
			w.logger.Printf("parent context is close")
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
			// If the channel is open and a task is received successfully:
			if ok {
				// Update the worker's status to indicate that it is currently running a task.
				// The `StatusWorkerRunning` status reflects that the worker is actively processing a job.
				w.setStatus(worker.StatusWorkerRunning)

				// Assign the received task to the worker's `currentProcess` field.
				// This keeps track of the task currently being executed by the worker.
				w.currentProcess = task

				// Increment the WaitGroup counter to account for the new task being processed.
				// This helps synchronize the completion of the task with other concurrent operations.
				wg.Add(1)

				// Set the WaitGroup for the task. This allows the task to signal when it has completed.
				// The `task.SetWaitGroup(wg)` call ensures that the task can signal its completion.
				_ = task.SetWaitGroup(wg)

				// Execute the task's `Run` method.
				// This method contains the logic to process the task.
				go task.Run()

				// After the task completes, reset the worker's `currentProcess` to `nil`.
				// This clears the reference to the completed task.
				w.currentProcess = nil

				// Update the worker's status to indicate that it is idle and ready for the next task.
				// The `StatusWorkerIdle` status reflects that the worker has finished the current task
				// and is available for new work.
				w.setStatus(worker.StatusWorkerIdle)
			} else {
				// If the task queue channel is closed
				w.logger.Printf("job collector is close: workerName: %s, workerStatus: %d", w.workerName, w.status)
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

// setStatus is responsible for safely updating the status of a Worker instance in a concurrent environment.
// It uses a mutex to ensure that only one goroutine can modify the worker's status at a time, preventing
// race conditions that could lead to inconsistent state or unexpected behavior. The method locks the worker's
// state, sets the status, and then releases the lock, ensuring that the operation is thread-safe.
func (w *Worker) setStatus(status worker.Status) {
	// Lock the worker's mutex to protect its state from concurrent access.
	// This ensures that the worker's status is set safely without race conditions.
	w.mutex.RLock()

	// Ensure that the mutex is unlocked when the function exits.
	// This is done using a deferred call to RUnlock, which will execute after the function returns.
	defer w.mutex.RUnlock()

	// Set the worker's status to the provided status value.
	// This updates the worker's status field with the new status.
	w.status = status
}

// Stop signals the worker to stop processing tasks and returns a channel to indicate completion.
// It closes the stop channel, causing the worker to exit its processing loop and finish the current job.
// If there is a task in processing at the time of worker termination, it will be stopped.
func (w *Worker) Stop() <-chan struct{} {
	defer func() {
		// Attempt to recover from a panic and retrieve the error.
		if rec := recover(); rec != nil {
			// Convert the recovered value to an error.
			if err := worker.GetRecoverError(rec); err != nil {
				// This is in case something caused a panic, but the worker status was not set to worker.StatusWorkerStopped,
				// so that the Worker pool would not recover the worker.
				if w.GetStatus() != worker.StatusWorkerStopped {
					w.setStatus(worker.StatusWorkerStopped)
				}
			}
		}
	}()

	// Create a channel to signal when the worker has stopped.
	// The buffered channel allows sending a single signal indicating the stop process is complete.
	doneCh := make(chan struct{}, 1)

	// Ensure the stop process is only executed once using sync.Once.
	w.onceStop.Do(func() {
		// Acquire a read lock on the worker to prevent concurrent shutdown operations.
		w.mutex.RLock()
		// Release the read lock after the shutdown sequence is complete.
		defer w.mutex.RUnlock()

		// Set the worker's status to "stopped" to indicate that it is no longer active.
		w.setStatus(worker.StatusWorkerStopped)

		// If there is a task currently being processed, stop it.
		// This ensures that any ongoing work is properly terminated.
		if w.currentProcess != nil {
			w.currentProcess.Stop()
		}

		// Send a signal through the done channel to indicate that the worker has stopped.
		doneCh <- struct{}{}
		// Close the done channel to indicate that no more signals will be sent.
		close(doneCh)

		// Send a signal through the stop channel to indicate the worker should stop processing.
		w.stopCh <- struct{}{}
		// Close the stop channel to indicate that no more stop signals will be sent.
		close(w.stopCh)
	})

	// Return the channel to allow external monitoring of completion.
	return doneCh
}

// GetStatus is a method that retrieves the current status of a Worker instance.
// It ensures that the status is accessed in a thread-safe manner by locking the
// worker's mutex before reading the status. This prevents race conditions and
// ensures that the value returned is consistent, even in a concurrent environment.
func (w *Worker) GetStatus() worker.Status {
	// Lock the worker's mutex to protect its state from concurrent access.
	// This ensures that the status is read safely without race conditions.
	w.mutex.RLock()

	// Ensure that the mutex is unlocked when the function exits.
	// This is done using a deferred call to RUnlock, which will execute after the function returns.
	defer w.mutex.RUnlock()

	// Return the worker's current status.
	// This provides the caller with the current value of the worker's status field.
	return w.status
}

// GetError returns the channel through which worker errors are communicated.
// This allows external components to listen for and handle errors generated by the worker.
// The channel is used to send instances of Error, containing information about the error and the worker instance.
// Note:In some cases I was able to get panic when running task. This is a very critical situation,
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

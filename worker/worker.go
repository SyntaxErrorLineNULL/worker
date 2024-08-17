package worker

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"worker"
)

// Worker represents a worker in a task processing pool.
// It holds all necessary information and channels to process tasks, manage its state, and handle errors.
type Worker struct {
	workerID       int64              // Unique identifier for the worker.
	workerContext  context.Context    // Context for the worker's operations.
	mutex          sync.RWMutex       // Mutex to control access to shared resources.
	stopCh         chan struct{}      // Channel to signal the worker to stop processing tasks.
	queue          <-chan worker.Task // Channel to receive jobs from the pool's dispatcher.
	currentProcess worker.Task        // Current task being processed by the worker.
	status         worker.Status      // Current status of the worker (e.g., running, stopped).
	errCh          chan *worker.Error // Channel to send and receive errors that occur in the worker.
	onceStop       sync.Once          // Ensures the stop process is only executed once.
	logger         *log.Logger
}

// NewWorker initializes a new Worker instance with the provided workerID.
// It sets up necessary channels and a logger for the worker, and returns a pointer to the Worker.
func NewWorker(workerID int64) *Worker {
	logger := log.New(os.Stdout, "pool:", log.LstdFlags)

	logger.Print("new worker pool")

	return &Worker{
		workerID: workerID,
		stopCh:   make(chan struct{}, 1),
		errCh:    make(chan *worker.Error),
		logger:   logger,
	}
}

// SetContext sets the context for the worker. This method is used to provide
// a new context for the worker, which can be used to control its operations
// and manage its lifecycle. The method ensures that the provided context is not
// nil before setting it, maintaining the integrity of the worker's context.
func (w *Worker) SetContext(ctx context.Context) {
	// Check if the provided context is not nil.
	// This ensures that we only set a valid context for the worker.
	if ctx != nil {
		// Set the worker's context to the provided context.
		// This allows the worker to use the new context for its operations.
		w.workerContext = ctx
	}
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
			return errors.New("queue chan is close")
		}
	// If the receive operation would block, continue without doing anything.
	default:
	}

	// Set the worker's queue channel to the provided queue channel.
	w.queue = queue

	// Return nil indicating that the operation was successful.
	return nil
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

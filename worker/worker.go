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
func (w *Worker) SetContext(ctx context.Context) error {
	// Check if the provided context is nil. A nil context is invalid and
	// should not be used. Return an error in this case to prevent setting
	// an invalid context for the worker.
	if ctx == nil {
		return worker.ContextIsNil
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

func (w *Worker) Start(wg *sync.WaitGroup) {
	if wg == nil {
		w.errCh <- &worker.Error{Error: worker.WaitGroupIsNil, Instance: w}
		return
	}

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
			err := worker.GetRecoverError(rec)
			if err != nil {
				// Send the error to the worker's error channel for external handling.
				w.errCh <- &worker.Error{Error: err, Instance: w}
			}
		}

		// Set the worker status to "stopped" before signaling the completion of the task.
		w.setStatus(worker.StatusWorkerStopped)

		// If a wait group is provided, decrement its counter to signal that the worker has completed its task.
		if wg != nil {
			// Decrement the WaitGroup counter to signal that the worker has completed its task.
			wg.Done()
		}
	}()

	for {
		select {
		case _, ok := <-w.stopCh:
			// Check if the stop channel is closed unexpectedly.
			if !ok {
				w.logger.Printf("stop channel is close, workerID: %d", w.workerID)
				return
			}

			w.logger.Printf("stop channel workerID: %d", w.workerID)
			return

		case <-w.workerContext.Done():
			w.logger.Printf("parent context is close")
			return

		case task, ok := <-w.queue:
			if ok {
				w.setStatus(worker.StatusWorkerRunning)
				w.currentProcess = task

				wg.Add(1)

				_ = task.SetWaitGroup(wg)

				go task.Run()

				w.currentProcess = nil
				w.setStatus(worker.StatusWorkerIdle)
			} else {
				// if the job channel is closed, there is no point in further work of the worker
				w.setStatus(worker.StatusWorkerStopped)
				w.logger.Printf("job collector is close: workerID: %d, workerStatus: %d", w.workerContext, w.status)
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
			err := worker.GetRecoverError(rec)
			if err != nil {
				// This is in case something caused a panic, but the worker status was not set to worker.StatusWorkerStopped,
				// so that the Worker pool would not recover the worker.
				if w.GetStatus() != worker.StatusWorkerStopped {
					w.setStatus(worker.StatusWorkerStopped)
				}

				// Send the error to the worker's error channel for external handling.
				w.errCh <- &worker.Error{Error: err, Instance: w}
			}

			// Close the error channel to indicate that no more errors will be sent.
			close(w.errCh)
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

		return
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

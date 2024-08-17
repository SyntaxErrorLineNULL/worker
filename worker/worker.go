package worker

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"worker"
)

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

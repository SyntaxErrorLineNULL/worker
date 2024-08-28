package worker

import (
	"context"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/SyntaxErrorLineNULL/worker"
)

type Pool struct {
	ctx               context.Context    // Parent context for the pool.
	contextCancelFunc context.CancelFunc // Cancel function for the pool's context.
	taskQueue         chan worker.Task   // Channel for queueing task's.
	workers           []worker.Worker    // Slice to hold worker instances.
	maxWorkersCount   int32              // Maximum number of workers in the pool.
	workerConcurrency atomic.Int32       // Number of currently running workers.
	poolWg            sync.WaitGroup     // Wait group for tracking running workers
	workerWg          sync.WaitGroup
	stopCh            chan struct{}      // Channel to signal stopping the pool.
	mutex             sync.Mutex         // Mutex for locking access to the pool.
	onceStart         sync.Once          // Used for a one-time action (starting the pool).
	onceStop          sync.Once          // Used for a one-time action (stopping the pool).
	stopped           bool               // flag signaling that the worker pool has already been stopped. The flag is needed to understand whether it is necessary to restore some worker after its fall, perhaps the worker pool is already stopped.
	workerErrorCh     chan *worker.Error // A channel for worker error. It is needed so that in case of panic we can restore the Worker's operation.
	logger            *log.Logger
}

func NewWorkerPool(options *worker.Options) *Pool {
	logger := log.New(os.Stdout, "pool:", log.LstdFlags)

	logger.Print("new worker pool")

	concurrency := options.WorkerCount

	if concurrency == 0 {
		concurrency = int32(runtime.NumCPU() * 2)
	}

	ctx, cancel := context.WithCancel(options.Context)

	return &Pool{
		ctx:               ctx,
		contextCancelFunc: cancel,
		taskQueue:         options.Queue,
		workers:           make([]worker.Worker, 0, options.WorkerCount),
		maxWorkersCount:   concurrency,
		stopCh:            make(chan struct{}, 1),
		stopped:           false,
		workerErrorCh:     make(chan *worker.Error),
		logger:            logger,
	}
}

// AddTaskInQueue attempts to add a task to the pool's task queue for processing.
// It performs several safety checks, including recovering from potential panics
// and ensuring the queue is valid before adding the task. If the queue is not
// initialized or has been closed, appropriate errors are returned.
func (p *Pool) AddTaskInQueue(task worker.Task) (err error) {
	// Use a defer statement to recover from panics and log any errors
	defer func() {
		if rec := recover(); rec != nil {
			// Convert the recovered panic value into an error.
			// This ensures that any panic during task addition is properly handled.
			err = worker.GetRecoverError(rec)
			// If an error is successfully created from the panic, return immediately.
			// This prevents further execution in case of a critical failure.
			if err != nil {
				return
			}
		}
	}()

	// Check if the taskQueue is nil, indicating that the queue has not been initialized.
	// If the taskQueue is nil, return an error indicating that the channel is empty.
	if p.taskQueue == nil {
		return worker.ChanIsEmpty
	}

	// Use a select statement to check the state of the taskQueue channel.
	// The purpose of this check is to see if the taskQueue channel has been closed.
	select {
	case <-p.taskQueue:
		// If the channel is closed, return an error indicating that the channel is closed.
		// This prevents adding tasks to a closed channel, which would cause a panic.
		return worker.ChanIsClose
	default:
		// If the channel is not closed, continue execution without blocking.
		// The default case allows the program to move on to adding the task to the queue.
	}

	// Add the provided task to the taskQueue for processing.
	// This operation is non-blocking and will place the task in the queue to be picked up by a worker.
	p.taskQueue <- task

	// Return nil to indicate that the task was successfully added to the queue.
	return nil
}

// RunningWorkers returns the current number of running workers in the pool.
// It retrieves the count of active workers using atomic operations to ensure thread safety.
func (p *Pool) RunningWorkers() int32 {
	// Return the current value of workerConcurrency, which represents the number of running workers.
	// The Load method is used to safely read the value atomically.
	return p.workerConcurrency.Load()
}

// incrementWorkerCount attempts to increment the worker count if it's below the maximum limit.
// It protects the worker count and associated wait group using a mutex to ensure
// thread-safety while managing the pool's worker count.
// If the current worker count is less than the maximum allowed workers, it increments
// the worker count and adds one to the wait group, signifying that a new worker is being started.
// If the maximum worker limit has been reached, it returns false to indicate that no more workers can be started.
func (p *Pool) incrementWorkerCount() bool {
	// Lock the mutex to protect the worker count and wait group.
	// This prevents race conditions when modifying the worker count.
	p.mutex.Lock()
	// Ensure the mutex is unlocked when the function returns.
	defer p.mutex.Unlock()

	// Get the current count of running workers.
	counter := p.RunningWorkers()

	// Check if the current worker count has reached the maximum limit.
	if counter >= p.maxWorkersCount {
		// The maximum worker limit has been reached, no more workers can be started.
		return false
	}

	// Increment the worker counter.
	p.workerConcurrency.Add(1)

	// Return true to indicate that the worker count was successfully incremented.
	return true
}

// decrementWorkerCount decrements the worker count in the worker pool.
// It utilizes an atomic operation to safely decrease the worker count.
// This method is called when a worker is stopped or removed from the pool.
func (p *Pool) decrementWorkerCount() {
	// Use atomic operation to safely decrease the worker count.
	// atomic.AddInt32(&p.workerConcurrency, -1)
	p.workerConcurrency.Add(-1)
}

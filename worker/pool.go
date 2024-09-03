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
	workerWg          *sync.WaitGroup
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
		workerWg:          new(sync.WaitGroup),
		logger:            logger,
	}
}

// Run starts the worker pool and worker goroutines. It creates and launches a specified number of worker goroutines,
// each of which is responsible for processing jobs from a shared collector. This method also continuously listens for stop signals
// or context cancellation and reacts accordingly, ensuring a clean and controlled shutdown of the worker pool.
// For safety once is used, this is done in case someone will use worker pool in more than one place,
// and it will happen that Start will be started again, nothing critical will happen and we will not lose the past handlers.
func (p *Pool) Run() {
	p.onceStart.Do(func() {
		// Add one to the wait group for the main pool loop.
		p.workerWg.Add(1)
		// Start the main pool loop in a goroutine.
		go p.loop()
	})
}

func (p *Pool) loop() {
	defer func() {
		// Recover from any panic in the job and report it.
		if rec := recover(); rec != nil {
			err := worker.GetRecoverError(rec)
			if err != nil {
				return
			}
		}

		p.workerWg.Done()
		p.contextCancelFunc()
	}()

	for {
		select {
		case <-p.stopCh:
			p.logger.Println("stop pool")
			return
		case <-p.ctx.Done():
			if !p.stopped {
				p.logger.Println("stop pool")
				// TODO: call Stop()
			}

			// Exit the loop.
			return
		case workerError, ok := <-p.workerErrorCh:
			if !ok {
				break
			}

			if !p.stopped {
				p.workerWg.Add(1)
				// it is safe to add waiting, because in case of panic triggering,
				// the waitGroup counter will be decreased afterwards.
				go workerError.Instance.Start(p.workerWg)
			}
		}
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
		return worker.ChanIsEmptyError
	}

	// Use a select statement to check the state of the taskQueue channel.
	// The purpose of this check is to see if the taskQueue channel has been closed.
	select {
	case <-p.taskQueue:
		// If the channel is closed, return an error indicating that the channel is closed.
		// This prevents adding tasks to a closed channel, which would cause a panic.
		return worker.ChanIsCloseError
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

// AddWorker adds a new worker to the worker pool and starts its execution.
// It ensures that the worker is correctly initialized with the pool's context
// and task queue, and handles any errors or panics that occur during the process.
func (p *Pool) AddWorker(wr worker.Worker) (err error) {
	// Check if the pool has been stopped. If it has, return an error indicating
	// that no more workers can be added.
	if p.stopped {
		return worker.WorkerPoolStopError
	}

	// Check if the provided worker is nil. If it is, return an error indicating
	// that a nil worker cannot be added.
	if wr == nil {
		return worker.WorkerIsNilError
	}

	// Use a defer statement to recover from any panic that occurs during the
	// addition of the worker and convert it into an error.
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

	// Attempt to increment the worker count. This checks if adding another worker
	// would exceed the maximum allowed workers in the pool.
	// If the worker count cannot be incremented (e.g., because the limit has been reached),
	// return an error indicating that the maximum number of workers has been reached.
	if !p.incrementWorkerCount() {
		return worker.MaxWorkersReachedError
	}

	// Lock the pool's mutex to ensure thread safety when modifying the pool's state.
	// This is necessary because multiple goroutines could attempt to add workers simultaneously.
	p.mutex.Lock()
	// Ensure the mutex is unlocked when the function returns to avoid deadlocks.
	defer p.mutex.Unlock()

	// Set the context for the worker. This context is used to control the worker's
	// execution, including cancellation and timeouts. If setting the context fails,
	// return the encountered error.
	if err = wr.SetContext(p.ctx); err != nil {
		return err
	}

	// Assign the task queue to the worker. The worker will pull tasks from this queue
	// for processing. If setting the queue fails, return the encountered error.
	if err = wr.SetQueue(p.taskQueue); err != nil {
		return err
	}

	// Append the worker to the pool's slice of workers.
	// This adds the worker to the internal tracking structure of the pool.
	p.workers = append(p.workers, wr)

	// Increment the WaitGroup counter to track this worker's lifecycle.
	// This ensures that the pool can wait for all workers to complete before shutting down.
	p.workerWg.Add(1)

	// Start the worker in a new goroutine.
	// The worker will begin processing tasks from the queue immediately.
	// The worker's lifecycle is tracked using the WaitGroup.
	go wr.Start(p.workerWg)

	// Return nil to indicate that the worker was successfully added and started.
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

// Stop terminates the worker pool and all its associated workers.
// It ensures a clean shutdown by canceling the pool's context, signaling
// workers to stop, and waiting for all workers to finish before completing
// the shutdown sequence.
func (p *Pool) Stop() {
	// Defer a function to recover from any panics that occur during the shutdown process.
	// This ensures that if something goes wrong, the panic is logged, and the shutdown continues.
	defer func() {
		if r := recover(); r != nil {
			p.logger.Println("stop pool panic")
			return
		}
	}()

	// Use sync.Once to ensure that the shutdown process is only executed once.
	// This prevents the pool from being shut down multiple times, which could lead to errors.
	p.onceStop.Do(func() {
		// Lock the mutex to ensure that the shutdown process is thread-safe.
		// This prevents other goroutines from interfering with the shutdown process.
		p.mutex.Lock()
		// Defer unlocking the mutex to ensure that it is always released, even if the shutdown fails.
		defer p.mutex.Unlock()

		// Call workersShutdown to stop all active workers gracefully.
		// This method will ensure that each worker completes its current task before shutting down.
		p.workersShutdown()

		// Set the workers slice to nil to release the memory and indicate that the pool no longer has any workers.
		p.workers = nil

		// Cancel the pool's context, which will signal to any remaining tasks that they should stop.
		// This is useful for stopping any long-running tasks that are still in progress.
		p.contextCancelFunc()

		// Mark the pool as stopped by setting the stopped flag to true.
		// This prevents any new tasks from being submitted to the pool.
		p.stopped = true

		// Send a signal to the stopCh channel to indicate that the stop process has started.
		// This can be used by other goroutines to wait for the shutdown to complete.
		p.stopCh <- struct{}{}
		// Close the stopCh channel to prevent any further sends on this channel.
		// This also signals to any listeners that the stop process is complete.
		close(p.stopCh)

		// Return from the Do function. Since this is inside sync.Once, this code block will not be executed again.
		return
	})

	// Wait for all workers to finish processing their current tasks before returning from the Stop method.
	// This ensures that all tasks are completed before the pool is fully shut down.
	p.workerWg.Wait()

	// Return from the Stop method, indicating that the pool has been successfully stopped.
	return
}

// workersShutdown handles the shutdown process for all workers in the pool.
// It ensures that each worker is properly stopped and decrements the worker count accordingly.
// This method recovers from any panic that occurs during the shutdown to avoid crashing the program.
func (p *Pool) workersShutdown() <-chan struct{} {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	// Create a channel to signal the completion of the shutdown process.
	// This channel will be closed once all workers have stopped.
	doneCh := make(chan struct{}, 1)

	// Create a slice to hold the done channels of all workers.
	// This slice will be used to wait for all workers to finish shutting down.
	doneChs := make([]<-chan struct{}, 0, p.workerConcurrency.Load())

	// Iterate over all workers in the pool.
	for _, wr := range p.workers {
		// Stop each wr that is not already stopped.
		if wr.GetStatus() != worker.StatusWorkerStopped {
			// If the worker is active, initiate its shutdown process by calling its Stop method.
			// Collect the worker's done channel in the slice for later synchronization.
			doneChs = append(doneChs, wr.Stop())
		} else {
			// If the worker is already stopped (e.g., due to context cancellation),
			// decrement the worker count as this worker will no longer be active.
			p.decrementWorkerCount()
		}
	}

	// Wait for all workers to finish their shutdown process.
	for _, ch := range doneChs {
		// Block until each worker's done channel is closed, indicating that the worker has stopped.
		<-ch
		// Decrement the worker count after each worker completes its shutdown.
		p.decrementWorkerCount()
	}

	// Send a signal on the done channel to indicate that the shutdown process is complete.
	// This allows external entities to know when all workers have been successfully stopped.
	doneCh <- struct{}{}
	// Close the done channel to signal completion and ensure no further sends can occur.
	close(doneCh)

	// Return the done channel to allow external monitoring of the shutdown completion.
	return doneCh
}

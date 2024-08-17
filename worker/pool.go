package worker

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"worker"
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
}

func NewWorkerPool(options *worker.Options) *Pool {
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
	}
}

package worker

// Pool defines the core interface for managing a pool of workers that execute tasks concurrently.
// It provides methods to run the pool, add tasks, manage workers, and control the pool's lifecycle.
type Pool interface {
	// Run starts the pool and initializes the worker management loop.
	// This function is responsible for handling task processing, managing worker errors, and ensuring
	// the lifecycle of workers, including graceful shutdowns when required.
	Run()

	// AddTaskInQueue attempts to add a new task to the pool's task queue for processing by the workers.
	// If the task cannot be added due to the pool being closed or other reasons, an error is returned.
	// This method ensures that the task is handled in a non-blocking way and leverages context-based
	// cancellation or timeout handling if needed.
	AddTaskInQueue(task Task) error

	// AddWorker registers a new worker to the pool. It initializes the worker with the pool's context,
	// task queue, and error handling mechanisms. The worker begins processing tasks after being added.
	// If the pool has been stopped or the worker cannot be added, an error is returned.
	AddWorker(wr Worker) error

	// RunningWorkers returns the number of currently active workers in the pool.
	// This value is managed atomically to ensure accurate results, even in a concurrent environment.
	RunningWorkers() int32

	// Stop gracefully shuts down the pool, signaling all workers to stop processing tasks.
	// It ensures all workers finish their ongoing tasks before the pool is fully shut down.
	Stop()
}

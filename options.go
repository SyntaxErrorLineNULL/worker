package worker

import "context"

// Options defines the configuration settings for a worker pool. It contains
// the context for task processing, the task queue, and settings related to
// the number of workers and retry behavior.
type Options struct {
	// Context is the context used to control the lifecycle of the tasks and workers.
	// It allows for task cancellation, timeouts, and passing values between functions.
	Context context.Context

	// Queue is the channel where tasks are submitted to be processed by the workers.
	// The size of the channel buffer determines how many tasks can be queued before
	// workers are required to start processing them.
	//
	// Note: the channel must be closed after the worker pool is stopped.
	Queue chan Task
	// WorkerCount specifies the number of workers that will be spawned in the pool.
	// This determines how many tasks can be processed concurrently.
	WorkerCount int32
	// MaxRetryWorkerRestart defines the maximum number of times a worker can be restarted
	// in case of failure before it is considered as a critical issue and further restarts
	// are stopped. This helps in preventing endless restarts in case of persistent errors.
	MaxRetryWorkerRestart int32
}

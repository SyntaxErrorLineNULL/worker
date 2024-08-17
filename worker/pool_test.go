package worker

import (
	"context"
	"testing"
	"worker"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Parallel()

	// InitWorkerPool tests the initialization of a worker pool with a specified number of workers
	// It ensures that the worker pool is correctly set up with the right worker count, parent context, and logger,
	// and verifies that no workers are running initially.
	t.Run("InitWorkerPool", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(16)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the specified parameters.
		pool := NewWorkerPool(&worker.Options{
			Context:     parentCtx,
			Queue:       task,
			WorkerCount: workerCount,
		})

		// Assert that initially, no workers should be running.
		// This ensures that the worker pool starts in an idle state before jobs are submitted.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Assert that the maximum number of workers in the pool matches the expected value.
		// This verifies that the worker pool was initialized with the correct number of workers.
		assert.Equal(t, pool.maxWorkersCount, workerCount, "Max worker count should match the specified worker count")

		// Assert that the pool is not in a stopped state initially.
		// This ensures that the pool is active and ready to process jobs.
		assert.False(t, pool.stopped, "Pool should not be marked as stopped initially")
	})
}

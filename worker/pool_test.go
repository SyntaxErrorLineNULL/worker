package worker

import (
	"context"
	"github.com/SyntaxErrorLineNULL/worker/mocks"
	"testing"

	"github.com/SyntaxErrorLineNULL/worker"
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

	// WorkerShutdown tests the behavior of the worker pool's workerShutdown method.
	// It verifies that the method properly shuts down a worker, reduces the number
	// of active workers, and removes the worker from the worker pool.
	// Additionally, it ensures that the proper workers remain after shutdown.
	t.Run("WorkerShutdown", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(16)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the specified options.
		// The pool will use the parent context, the task queue, and the defined number of workers.
		pool := NewWorkerPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

		// Assert that initially, no workers should be running.
		// This ensures that the worker pool starts in an idle state before jobs are submitted.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Increment the worker concurrency count by 5.
		// This sets the internal concurrency counter to 5 workers.
		pool.workerConcurrency.Add(5)

		// Create a mock worker to simulate a real worker for the test.
		// The mock allows us to control and observe the worker's behavior.
		firstMockWorker := mocks.NewWorker(t)
		// Create a second mock worker to simulate another real worker.
		// Multiple mock workers allow testing of worker shutdown behavior.
		secondMockWorker := mocks.NewWorker(t)
		// Create a third mock worker that will be the one shut down during the test.
		// This worker will be specifically tested for proper shutdown handling.
		lastMockWorker := mocks.NewWorker(t)

		// Expect the last mock worker to stop, and return nil to indicate no errors.
		// This sets up the expectation that this worker will be stopped during the test.
		lastMockWorker.EXPECT().Stop().Return(nil).Once()

		// Define a slice to store the mock workers for the worker pool.
		// The slice holds all the workers to be used in the pool.
		var workers []worker.Worker
		// Add the three mock workers to the slice of workers.
		// This ensures that the pool has three workers ready for testing.
		workers = append(workers, firstMockWorker, secondMockWorker, lastMockWorker)

		// Assign the list of mock workers to the worker pool's workers.
		// This step sets up the pool with the mock workers that will be used in the test,
		// allowing the pool to manage and interact with these workers.
		pool.workers = workers
		// Assert that the number of workers in the pool matches the number of mock workers provided.
		// This verification ensures that the pool's workers list has been correctly updated with the mock workers.
		assert.Equal(t, len(workers), len(pool.workers), "Expected the pool to have the correct number of workers")

		// Call the workerShutdown method to shut down the last mock worker.
		// This method is expected to remove the worker from the pool and perform any necessary cleanup.
		pool.workerShutdown(lastMockWorker)

		// Assert that the number of running workers is now 4 after shutting down one worker.
		// This checks that the pool correctly updates the count of active workers following the shutdown process.
		assert.Equal(t, int32(4), pool.RunningWorkers(), "Expected 4 running workers after shutting down one worker")

		// Assert that the number of workers remaining in the pool is 2 after shutting down one worker.
		// This verifies that the pool correctly reflects the reduced number of workers.
		assert.Equal(t, 2, len(pool.workers), "Expected 2 workers to remain in the pool after shutdown")

		// Assert that the first worker in the pool is the first mock worker.
		// This ensures that the worker list order is maintained correctly after the shutdown operation.
		assert.Equal(t, firstMockWorker, pool.workers[0], "Expected the first worker in the pool to be firstMockWorker")

		// Assert that the second worker in the pool is the second mock worker.
		// This checks that the remaining workers are correctly ordered in the pool after removing the last worker.
		assert.Equal(t, secondMockWorker, pool.workers[1], "Expected the second worker in the pool to be secondMockWorker")
	})
}

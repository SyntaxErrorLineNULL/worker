package worker

import (
	"context"
	"github.com/SyntaxErrorLineNULL/worker"
	"github.com/SyntaxErrorLineNULL/worker/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPool(t *testing.T) {
	t.Parallel()

	// Create a new mock task instance using the mocks package.
	// The mock task is created for use in testing, providing a controlled and
	// predictable task object that can be configured to simulate various behaviors.
	// This is essential for verifying that the worker pool handles tasks as expected.
	mockTask := mocks.NewTask(t)
	// Assert that the mock task instance is not nil.
	// This ensures that the mock task was successfully created and initialized.
	// If the task were nil, it would indicate a failure in the mock setup, which
	// could lead to incorrect test results or an inability to run the test as intended.
	assert.NotNil(t, mockTask, "Mock task should not be nil")

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

	// AddWorkerWithNilWorker tests the behavior of the AddWorker method when
	// attempting to add a nil worker to the worker pool. This test ensures that
	// the method correctly identifies and handles invalid input. Specifically,
	// it checks whether the pool returns the appropriate error when a nil
	// worker is provided, and verifies that no changes occur in the pool state.
	t.Run("AddWorkerWithNilWorker", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(1)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewWorkerPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the worker pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Attempt to add a nil worker to the pool.
		// This tests the pool's handling of invalid input. Adding a nil worker is not
		// allowed, and the pool should properly reject it and return an error.
		err := pool.AddWorker(nil)

		// Assert that an error is returned when trying to add a nil worker.
		// This ensures that the pool correctly handles invalid input by returning an appropriate error.
		assert.Error(t, err, "Expected error when adding a nil worker, but no error was returned")

		// Assert that the error returned is specifically of type WorkerIsNilError.
		// This checks that the error type matches the expected error for adding a nil worker.
		assert.ErrorIs(t, err, worker.WorkerIsNilError, "The error returned when adding a nil worker was not of type WorkerIsNilError")
	})

	// AddWorkerWithStoppedPool tests the behavior of the AddWorker method when
	// attempting to add a worker to a pool that has already been stopped. This test
	// ensures that the pool correctly handles the scenario where worker addition is
	// attempted after the pool has been stopped. It verifies that the pool returns
	// the appropriate error, reflecting its stopped state and preventing any changes
	// to the pool’s worker list.
	t.Run("AddWorkerWithStoppedPool", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(1)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the given parameters.
		// The pool is set up with the provided context, task queue, and worker count.
		// This prepares the pool for task processing with the specified number of workers.
		pool := NewWorkerPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

		// Manually set the pool's stopped flag to true.
		// This simulates the condition where the pool has been stopped and is no longer
		// accepting new workers. This setup is necessary to test the pool's behavior
		// when in a stopped state.
		pool.stopped = true

		// Attempt to add a nil worker to the stopped pool.
		// This tests how the pool handles worker addition when it is in a stopped state.
		err := pool.AddWorker(nil)
		// Assert that an error is returned when attempting to add a worker to a stopped pool.
		// This confirms that the pool correctly identifies its stopped state and prevents
		// further modifications to its worker list.
		assert.Error(t, err, "Expected error when adding a worker to a stopped pool, but no error was returned")
		// Assert that the error returned is specifically of type WorkerPoolStopError.
		// This verifies that the pool returns the correct error type, indicating that
		// the pool is stopped and cannot accept new workers.
		assert.ErrorIs(t, err, worker.WorkerPoolStopError, "The error returned when adding a worker to a stopped pool was not of type WorkerPoolStopError")
	})

	// SuccessfullyAddsTask tests the behavior of the `AddTaskInQueue` method
	// when adding a task to the worker pool. It ensures that the task is
	// correctly added to the task queue and that no errors occur during
	// the addition process.
	t.Run("SuccessfullyAddsTask", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(1)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task, 1)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewWorkerPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the worker pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Expectation setup for mock task.
		// Mock task's SetContext method is expected to be called with the pool's context
		// and should return nil. This setup is necessary to simulate the behavior of
		// adding a task in the worker pool.
		mockTask.EXPECT().SetContext(pool.ctx).Return(nil).Once()

		// AddTaskInQueue adds a task to the worker pool's task queue.
		// This method call should enqueue the task and return an error if something goes wrong.
		// The task should be successfully added to the queue, as verified by the assertion below.
		err := pool.AddTaskInQueue(mockTask)
		// Assert that adding the task to the queue did not result in an error.
		// The assertion ensures that no error was returned during the operation,
		// which indicates that the task was successfully enqueued in the worker pool.
		assert.NoError(t, err, "Adding task to queue should not return an error")

		// Assert that the task was added to the collector channel by checking the length of the collector channel.
		// The length should be 1, indicating that the job is correctly queued in the pool.
		assert.Equal(t, 1, len(pool.taskQueue), "The queue channel should contain exactly one task")
	})

	// WorkerLimitReached tests the behavior of the worker pool when attempting to add
	// more workers than the maximum limit allowed by the pool. It verifies that the
	// pool enforces the worker limit correctly by allowing only up to the maximum
	// number of workers and rejecting any additional workers beyond this limit.
	t.Run("WorkerLimitReached", func(t *testing.T) {
		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(1)

		// Create a parent context for the worker pool.
		// The context is used to control the lifetime of the worker pool.
		parentCtx := context.Background()

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task, 1)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewWorkerPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the worker pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Add the first worker to the worker pool.
		// This should succeed since the pool's limit is 1 and currently no workers are running.
		err := pool.AddWorker(NewWorker("1"))
		// Check that adding the first worker does not produce an error.
		assert.NoError(t, err, "Adding the first worker should not produce an error")

		// Attempt to add a second worker to the worker pool.
		// Since the pool's limit is 1, adding another worker should fail.
		err = pool.AddWorker(NewWorker("2"))
		// Assert that an error occurred when trying to add the second worker.
		assert.Error(t, err, "Adding a second worker should produce an error as the worker limit is reached")
	})
}

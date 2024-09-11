package worker

import (
	"context"
	"errors"
	"github.com/SyntaxErrorLineNULL/worker"
	"github.com/SyntaxErrorLineNULL/worker/mocks"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Parallel()

	defer runtime.GC()

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

		// Assert that the pool's maximum worker restart limit is set to DefaultMaxRetry.
		// This confirms that if no custom retry value was provided, the pool defaults to the correct predefined constant.
		// Ensures that the retry behavior aligns with expected default behavior in the worker pool.
		assert.Equal(t, int32(worker.DefaultMaxRetry), pool.maxRetryWorkerRestart, "Expected default max retry value to be set")
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

	// WorkerPanic simulates a scenario where a worker in the worker pool encounters an error or panic condition.
	// The goal of this test is to validate how the worker pool responds when a worker fails unexpectedly.
	// Specifically, it ensures that the worker pool can correctly manage worker errors without crashing,
	// allows the worker to be added and managed even if it later encounters issues, and ensures that
	// the worker pool can continue processing other tasks normally. The test covers both the normal worker
	// lifecycle (such as restarting, starting, and queue handling) and the error handling mechanisms that
	// allow the pool to remain operational when individual workers fail. By simulating a worker panic,
	// this test ensures robustness in the pool's ability to manage errors while maintaining its ability
	// to continue task execution.
	t.Run("WorkerPanic", func(t *testing.T) {
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

		// Start the worker pool in a separate Goroutine to allow it to operate asynchronously.
		// This enables the pool to begin its job processing and worker management in parallel.
		go pool.Run()

		// Pause for 20 milliseconds to give the pool time to initialize and get ready.
		// This small delay ensures that the pool has started running before we add workers.
		<-time.After(20 * time.Millisecond)

		// Create a mock worker to simulate the behavior of a real worker in the pool.
		// The mock is used to define expected behaviors, such as restarting or starting a worker.
		newWorker := mocks.NewWorker(t)
		// Expect the mock worker to restart using the pool's wait group.
		// This simulates the worker being restarted, ensuring that the pool is managing worker lifecycle correctly.
		newWorker.EXPECT().Restart(pool.workerWg).Return()
		// Expect the mock worker to start, using the pool's wait group.
		// This simulates the worker starting its task processing after being added to the pool.
		newWorker.EXPECT().Start(pool.workerWg).Return()
		// Set the context for the worker, passing the pool's context.
		// This ensures that the worker is tied to the pool's context, allowing it to handle cancellation or timeouts.
		newWorker.EXPECT().SetContext(pool.ctx).Return(nil)
		// Set the task queue for the worker, using the pool's task queue.
		// This binds the worker to the queue from which it will pull tasks for processing.
		newWorker.EXPECT().SetQueue(pool.taskQueue).Return(nil)
		// Simulate the worker's retry behavior, returning a retry count of 1.
		// This ensures that the worker is allowed one retry if it encounters a failure during task execution.
		newWorker.EXPECT().GetRetry().Return(int32(1))
		// Set up the expected behavior for the worker's SetWorkerErrChannel method.
		// This expectation specifies that when the SetWorkerErrChannel method is called on the
		// newWorker mock, it should be provided with the workerErrorCh channel from the pool.
		newWorker.EXPECT().SetWorkerErrChannel(pool.workerErrorCh).Return(nil)

		// Add the mock worker to the pool and assert no error occurred.
		// This step verifies that the worker is correctly added to the pool.
		err := pool.AddWorker(newWorker)
		// Assert that adding the worker did not return an error.
		// This ensures that the worker addition process was successful and no issues occurred.
		assert.NoError(t, err, "Adding the worker to the pool should not produce an error")

		// Check that the worker was successfully added to the pool.
		// This verifies that the pool now contains one worker.
		assert.Equal(t, workerCount, int32(len(pool.workers)), "The worker pool should contain one worker")

		// Simulate an error occurring in the worker by sending an error to the worker's error channel.
		// This action allows the test to validate how the worker pool handles errors from workers.
		pool.workerErrorCh <- &worker.Error{Error: errors.New("some error"), Instance: newWorker}

		// Wait for 1 second to allow the job to be processed by the worker.
		// This provides sufficient time for the worker to process the job.
		<-time.After(100 * time.Millisecond)
	})

	// SuccessStop tests the behavior of stopping a worker pool and ensures that
	// all workers are properly terminated and the pool is in a stopped state.
	// It verifies that the pool correctly stops all workers, empties its worker list,
	// and sets appropriate statuses. The test also checks the worker's status after
	// the pool is stopped to ensure it is correctly updated.
	t.Run("SuccessStop", func(t *testing.T) {
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

		// Start the worker pool in a separate Goroutine to allow it to operate asynchronously.
		// This enables the pool to begin its job processing and worker management in parallel.
		go pool.Run()

		// Pause for 100 milliseconds to give the pool time to initialize and get ready.
		// This small delay ensures that the pool has started running before we add workers.
		<-time.After(100 * time.Millisecond)

		newWorker := NewWorker("test-stop-worker")
		// Add the new worker to the pool and assert no error occurred.
		// This step verifies that the worker is correctly added to the pool.
		err := pool.AddWorker(newWorker)
		// Assert that adding the worker did not return an error.
		// This ensures that the worker addition process was successful and no issues occurred.
		assert.NoError(t, err, "Adding the worker to the pool should not produce an error")

		// Assert that the number of running workers is equal to 1.
		// This ensures that the worker pool has exactly one worker running after adding the worker.
		assert.Equal(t, int32(1), pool.RunningWorkers(), "There should be exactly 1 worker running after adding the worker")

		// Retrieve the first worker instance from the pool's workers slice.
		// This step assumes that a worker has already been added to the pool.
		wr := pool.workers[0]

		// Ensure the retrieved worker instance is not nil.
		// This assertion checks that the worker was successfully created and added to the pool.
		// If the worker is nil, it indicates an issue with the worker creation or addition process.
		assert.NotNil(t, wr, "The worker instance should be successfully created")

		// Assert that the worker's current status is equal to StatusWorkerIdle.
		// This check ensures that the worker is in the expected idle state,
		// indicating that it is not currently processing any jobs and is ready
		// to take on new tasks.
		assert.Equal(t, worker.StatusWorkerIdle, wr.GetStatus())

		// Wait for 100 ms to allow the job to be processed by the worker.
		// This provides sufficient time for the worker to process the job.
		<-time.After(100 * time.Millisecond)

		// Create a channel to signal the completion of the pool stop operation.
		// This channel will be used to notify when the worker pool has been stopped.
		doneCh := make(chan struct{}, 1)

		// Start a Goroutine to stop the worker pool and signal completion via the done channel.
		go func() {
			// Ensure the channel is closed after stopping the pool.
			defer close(doneCh)
			// Stop the worker pool to terminate all workers.
			pool.Stop()
		}()

		// Wait for the pool to stop or timeout.
		// Check if the worker pool has stopped and no workers are running.
		select {
		case <-doneCh:
			// Check if the number of running workers is zero after stopping the pool.
			// This confirms that all workers have stopped as expected.
			assert.Equal(t, int32(0), pool.RunningWorkers(), "All workers should have stopped")

			// Check that the worker pool no longer contains any workers.
			// This verifies that the worker pool has been completely emptied after stopping.
			assert.Equal(t, 0, len(pool.workers), "The worker pool should be empty")

			// Verify that the worker pool has stopped.
			// The `pool.stopped` flag should be set to `true` if the pool was successfully
			// stopped as part of the test. This assertion ensures that the pool's stop logic
			// was executed correctly and that all worker goroutines were properly terminated.
			// The test will fail if `pool.stopped` is not `true`, indicating a potential issue
			// with the pool's stopping mechanism.
			assert.True(t, pool.stopped)

			// Check that the worker's status is set to `StatusWorkerStopped` after stopping.
			// This ensures that the worker was properly stopped as part of the pool shutdown process.
			assert.Equal(t, wr.GetStatus(), worker.StatusWorkerStopped, "The worker should have stopped")

		case <-time.After(6 * time.Second):
			// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
			t.Error("Failed to stop worker pool within expected time")
		}
	})

	// ContextDone tests the behavior of the worker pool when the parent context is canceled.
	// It verifies that the pool correctly stops all running workers, cleans up resources,
	// and updates its internal state to reflect that it has been stopped. This test ensures
	// proper shutdown of the pool and confirms that all workers are terminated as expected.
	t.Run("ContextDone", func(t *testing.T) {
		// Create a context with cancellation capabilities.
		// This context will be used to control the worker’s execution lifecycle,
		// allowing us to signal the worker to stop.
		ctx, cancel := context.WithCancel(context.Background())

		// Define the number of workers to be used in the worker pool.
		// This value determines how many worker goroutines will be created.
		workerCount := int32(1)

		// Create a channel for job submission.
		// Jobs will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task, 1)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewWorkerPool(&worker.Options{Context: ctx, Queue: task, WorkerCount: workerCount})

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the worker pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Start the worker pool in a separate Goroutine to allow it to operate asynchronously.
		// This enables the pool to begin its job processing and worker management in parallel.
		go pool.Run()

		// Pause for 100 milliseconds to give the pool time to initialize and get ready.
		// This small delay ensures that the pool has started running before we add workers.
		<-time.After(100 * time.Millisecond)

		newWorker := NewWorker("test-stop-worker")
		// Add the new worker to the pool and assert no error occurred.
		// This step verifies that the worker is correctly added to the pool.
		err := pool.AddWorker(newWorker)
		// Assert that adding the worker did not return an error.
		// This ensures that the worker addition process was successful and no issues occurred.
		assert.NoError(t, err, "Adding the worker to the pool should not produce an error")

		// Assert that the number of running workers is equal to 1.
		// This ensures that the worker pool has exactly one worker running after adding the worker.
		assert.Equal(t, workerCount, pool.RunningWorkers(), "There should be exactly 1 worker running after adding the worker")

		// Retrieve the first worker instance from the pool's workers slice.
		// This step assumes that a worker has already been added to the pool.
		wr := pool.workers[0]

		// Ensure the retrieved worker instance is not nil.
		// This assertion checks that the worker was successfully created and added to the pool.
		// If the worker is nil, it indicates an issue with the worker creation or addition process.
		assert.NotNil(t, wr, "The worker instance should be successfully created")

		// Assert that the worker's current status is equal to StatusWorkerIdle.
		// This check ensures that the worker is in the expected idle state,
		// indicating that it is not currently processing any jobs and is ready
		// to take on new tasks.
		assert.Equal(t, worker.StatusWorkerIdle, wr.GetStatus())

		// Pause for 500 milliseconds to allow the worker to remain idle for a brief period.
		// This gives the test enough time to ensure the worker is in the idle state
		// before the context is canceled and the shutdown process begins.
		<-time.After(500 * time.Millisecond)

		// Cancel the context to signal the pool to stop.
		cancel()

		// Wait for a short duration to allow the stop signal to propagate.
		<-time.After(100 * time.Millisecond)

		// Wait for the pool to stop or timeout.
		// Check if the worker pool has stopped and no workers are running.
		select {
		case <-pool.stopCh:
			// Check if the number of running workers is zero after stopping the pool.
			// This confirms that all workers have stopped as expected.
			assert.Equal(t, int32(0), pool.RunningWorkers(), "All workers should have stopped")

			// Check that the worker pool no longer contains any workers.
			// This verifies that the worker pool has been completely emptied after stopping.
			assert.Equal(t, 0, len(pool.workers), "The worker pool should be empty")

			// Verify that the worker pool has stopped.
			// The `pool.stopped` flag should be set to `true` if the pool was successfully
			// stopped as part of the test. This assertion ensures that the pool's stop logic
			// was executed correctly and that all worker goroutines were properly terminated.
			assert.True(t, pool.stopped, "Expected pool to be stopped")

			// Check that the worker's status is set to `StatusWorkerStopped` after stopping.
			// This ensures that the worker was properly stopped as part of the pool shutdown process.
			assert.Equal(t, wr.GetStatus(), worker.StatusWorkerStopped, "The worker should have stopped")

		case <-time.After(5 * time.Second):
			// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
			t.Error("Failed to stop worker pool within expected time")
		}
	})
}

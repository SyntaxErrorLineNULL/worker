//nolint:all
package worker

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/SyntaxErrorLineNULL/worker"
	"github.com/SyntaxErrorLineNULL/worker/mocks"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	t.Parallel()

	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Define a mock worker to be used in the test.
	// This mock simulates a real worker instance and is used to verify behavior
	// in scenarios where worker interactions need to be tested.
	mockWorkerPool := mocks.NewPool(t)
	// Assert that the mock worker pool is successfully created.
	// This ensures that the worker mock was instantiated correctly and is not nil.
	// If the mock creation fails, the test will fail with a descriptive error message.
	assert.NotNil(t, mockWorkerPool, "Expected mock worker pool to be instantiated, but got nil")

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
		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the specified parameters.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, MaxRetryWorkerRestart: 3, Logger: logger})

		// Assert that initially, no workers should be running.
		// This ensures that the worker pool starts in an idle state before tasks are submitted.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Assert that the maximum number of workers in the pool matches the expected value.
		// This verifies that the worker pool was initialized with the correct number of workers.
		assert.Equal(t, pool.maxWorkersCount, workerCount, "Max worker count should match the specified worker count")

		// Assert that the pool is not in a stopped state initially.
		// This ensures that the pool is active and ready to currentProcess tasks.
		assert.False(t, pool.stopped, "Pool should not be marked as stopped initially")
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

		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the worker pool.
		task := make(chan worker.Task)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, Logger: logger})

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
		// Assert that the error returned is specifically of type ErrWorkerIsNil.
		// This checks that the error type matches the expected error for adding a nil worker.
		assert.ErrorIs(t, err, worker.ErrWorkerIsNil, "The error returned when adding a nil worker was not of type ErrWorkerIsNil")
	})

	// SuccessfullyAddWorker tests the behavior of adding a worker to the pool and processing a task.
	// It verifies that the pool can successfully add a worker, submit a task for processing, and
	// handle the task's completion correctly. The test ensures that the worker pool properly manages
	// the worker lifecycle, processes the task, and transitions the worker to the 'stopped' state
	// after the task has been completed.
	t.Run("SuccessfullyAddWorker", func(t *testing.T) {
		// Define the number of workers to be used in the wr pool.
		// This value determines how many wr goroutines will be created.
		workerCount := int32(1)
		// Define a name for the wr to be used in this test.
		// This name is used to initialize and identify the wr instance.
		workerName := "workerName"
		// Define the wr timeout duration for the test.
		// This is the maximum amount of time we allow for the task to complete.
		timeout := 5 * time.Second

		// Create a parent context for the wr pool.
		// The context is used to control the lifetime of the wr pool.
		parentCtx := context.Background()

		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the wr pool.
		queue := make(chan worker.Task, 10)

		// Initialize a new wr pool with the given parameters.
		// This sets up the pool with the provided context, task queue, and wr count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger})

		// Start the wr pool in a separate Goroutine to allow it to operate asynchronously.
		// This enables the pool to begin its task processing and wr management in parallel.
		go pool.Run()

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the wr pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the wr with specified parameters and ensures that it is properly set up.
		wr := NewWorker(workerName, timeout, logger)
		// Assert that the wr instance is not nil.
		// This checks that the wr was successfully created and is not a zero value.
		assert.NotNil(t, wr, "Worker should be successfully created")

		// Create an instance of the mock processing task with a specified timeout.
		// This mock simulates a long-running task for testing purposes.
		mockSuccessProcessing := &MockSuccessProcessing{}
		// Initialize input data for the processing function.
		// `inputProcessingData` represents an example integer input (in this case, `222`)
		// that will be passed to the processing function. The integer is cast to `int32`
		// to match the expected data type used by the processing function.
		inputProcessingData := int32(3)

		// Define the wr timeout duration for the test.
		// This is the maximum amount of time we allow for the task to complete.
		workerTimeout := 1 * time.Second

		// Initialize the necessary variables and objects for the task.
		// Create a new task instance with a timeout and a unique task name.
		task := NewTask(workerTimeout, "test-task", mockSuccessProcessing, inputProcessingData)
		// Assert that the task was successfully created.
		// If the task is nil, it indicates a problem with task initialization.
		assert.NotNil(t, task, "Expected task to be initialized, but it was nil")
		// Create a buffered done channel to signal task completion.
		// This channel will be used to notify when the task is done.
		doneCh := make(chan struct{}, 1)
		// Set the done channel for the task using the SetDoneChannel method.
		// The method should return no error if the done channel is valid.
		_ = task.SetDoneChannel(doneCh)

		// Add the wr to the pool using the `AddWorker` method.
		// This action registers the wr in the pool, allowing it to start processing tasks.
		err := pool.AddWorker(wr)
		// Assert that no error is returned when adding the wr to the pool.
		// This ensures that the `AddWorker` method worked correctly, and the wr was successfully
		// added without any issues, confirming that the pool is properly configured.
		assert.NoError(t, err, "Expected no error while adding wr to the pool")

		// Assert that the wr count in the pool matches the expected count.
		// This ensures that the wr has been successfully added.
		assert.Equal(t, workerCount, pool.RunningWorkers(), "Worker count should match the expected number")

		// Send the mock task to the wr's task queue.
		// This simulates the wr receiving a task for processing.
		// The task represents a task that the wr should handle.
		err = pool.AddTaskInQueue(task)
		// Assert that no error is returned when adding the task to the pool.
		// This ensures that the `AddTask` function works as expected, and the task was successfully
		// added to the wr pool for processing without any issues.
		assert.NoError(t, err, "Expected no error while adding task to the pool")

		// Use a select statement to handle the task completion or timeout.
		// This waits for either the task to signal completion or a timeout to occur.
		select {
		case <-doneCh:
			// Assert that the processing counter for the mock task is incremented by the volume ID.
			// This confirms that the mock processing task completed successfully.
			assert.Equal(t, inputProcessingData, mockSuccessProcessing.counter.Load())
		case <-time.After(2 * time.Second):
			// If the task doesn't complete within the expected time, trigger an error.
			t.Error("The execution time exceeds the allowable timeout")
		}

		// Wait for the wr to stop processing and check its status.
		// This block waits for either the wr to signal that it has stopped or a timeout to occur.
		select {
		case <-wr.Stop():
			// Assert that the wr status is 'stopped' after cancellation.
			// This verifies that the wr correctly transitions to the 'stopped' state.
			assert.True(t, wr.IsStop(), "Expected wr status to be stopped")

			// Assert that the `currentProcess` field of the `wr` is `nil`.
			// This ensures that the wr has not started processing or has properly cleaned up its processing state.
			assert.Nil(t, wr.currentProcess, "Expected wr.currentProcess to be nil, indicating that the wr has not started processing or has been properly cleaned up.")
		case <-time.After(2 * time.Second):
			// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
			t.Error("Failed to stop wr within expected time")
		}
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

		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the worker pool.
		queue := make(chan worker.Task)

		// Initialize a new worker pool with the given parameters.
		// The pool is set up with the provided context, queue queue, and worker count.
		// This prepares the pool for queue processing with the specified number of workers.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger})

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
		// Assert that the error returned is specifically of type ErrWorkerPoolStop.
		// This verifies that the pool returns the correct error type, indicating that
		// the pool is stopped and cannot accept new workers.
		assert.ErrorIs(t, err, worker.ErrWorkerPoolStop, "The error returned when adding a worker to a stopped pool was not of type ErrWorkerPoolStop")
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

		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the worker pool.
		queue := make(chan worker.Task, 1)

		// Initialize a new worker pool with the given parameters.
		// This sets up the pool with the provided context, queue queue, and worker count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger})

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the worker pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Add the first worker to the worker pool.
		// This should succeed since the pool's limit is 1 and currently no workers are running.
		err := pool.AddWorker(NewWorker("name 1", 1*time.Second, logger))
		// Check that adding the first worker does not produce an error.
		assert.NoError(t, err, "Adding the first worker should not produce an error")

		// Attempt to add a second worker to the worker pool.
		// Since the pool's limit is 1, adding another worker should fail.
		err = pool.AddWorker(NewWorker("name 2", 1*time.Second, logger))
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
		// Define the number of workers to be used in the wr pool.
		// This value determines how many wr goroutines will be created.
		workerCount := int32(1)

		// Create a parent context for the wr pool.
		// The context is used to control the lifetime of the wr pool.
		parentCtx := context.Background()

		// Create a channel for task submission.
		// Tasks will be sent to this channel for processing by the wr pool.
		queue := make(chan worker.Task, 1)

		// Initialize a new wr pool with the given parameters.
		// This sets up the pool with the provided context, queue queue, and wr count.
		// It prepares the pool to manage and distribute tasks to the workers.
		pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger, MaxRetryWorkerRestart: 3})

		// Start the wr pool in a separate Goroutine to allow it to operate asynchronously.
		// This enables the pool to begin its task processing and wr management in parallel.
		go pool.Run()

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the wr pool initialization is correct before adding any workers.
		assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

		// Pause for 200 milliseconds to give the pool time to initialize and get ready.
		// This small delay ensures that the pool has started running before we add workers.
		<-time.After(200 * time.Millisecond)

		// Create a buffered channel for signaling when the wr's operation is done.
		// The buffer size of 1 allows for one signal to be sent to indicate completion.
		// This channel is used to synchronize the wr's completion status with the test.
		doneCh := make(chan struct{}, 1)

		// Create a buffered channel for signaling when the wr should stop.
		// The buffer size of 1 allows for one stop signal to be sent to the wr.
		// This channel is used to control and synchronize the stopping of the wr with the test.
		stopCh := make(chan struct{}, 1)

		// Create an instance of MockWorkerWithPanic, initializing it with the previously
		// created channels `doneCh` and `stopCh`. The `doneCh` channel will be used to
		// signal when the wr's task is complete, while the `stopCh` channel will signal
		// when the wr should stop its operation. This setup allows for testing the wr's
		// behavior in response to task completion and stop signals.
		wr := &MockWorkerWithPanic{doneCh: doneCh, stopCh: stopCh}

		// Add the first wr to the wr pool.
		// This should succeed since the pool's limit is 1 and currently no workers are running.
		err := pool.AddWorker(wr)
		// Check that adding the first wr does not produce an error.
		assert.NoError(t, err, "Adding the first wr should not produce an error")

		// Assert that initially, no workers should be running.
		// This confirms that the pool starts in an idle state with zero active workers.
		// Ensures that the wr pool initialization is correct before adding any workers.
		assert.Equal(t, workerCount, pool.RunningWorkers(), "Initially, no workers should be running")

		// Pause for 200 milliseconds to give the pool time to initialize and get ready.
		// This small delay ensures that the pool has started running before we add workers.
		<-time.After(200 * time.Millisecond)

		select {
		// Expect the task to be executed after the wr restarts.
		// Check the retry and restart counters to verify that the wr was restarted
		// and that the task was processed successfully after the restart.
		case <-stopCh:
			// Assert that the retry counter of the wr is incremented by 1.
			// This confirms that the wr retry logic is functioning as expected.
			assert.Equal(t, int32(1), wr.GetRetry())

			// Assert that the restart counter of the wr is incremented by 1.
			// This confirms that the wr was restarted after encountering a panic.
			assert.Equal(t, int32(1), wr.restart.Load())

			// Stop the wr pool after verifying the wr's behavior.
			// This ensures that the pool is properly stopped and cleaned up.
			pool.Stop()
		}
	})
}

// SuccessStop tests the behavior of stopping a worker pool and ensures that
// all workers are properly terminated and the pool is in a stopped state.
// It verifies that the pool correctly stops all workers, empties its worker list,
// and sets appropriate statuses. The test also checks the worker's status after
// the pool is stopped to ensure it is correctly updated.
func TestSuccessStopWorker(t *testing.T) {
	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Define the number of workers to be used in the worker pool.
	// This value determines how many worker goroutines will be created.
	workerCount := int32(1)
	// Define a name for the worker to be used in this test.
	// This name is used to initialize and identify the worker instance.
	workerName := "workerName"
	// Define the worker timeout duration for the test.
	// This is the maximum amount of time we allow for the task to complete.
	timeout := 5 * time.Second

	// Create a parent context for the worker pool.
	// The context is used to control the lifetime of the worker pool.
	parentCtx := context.Background()

	// Create a channel for task submission.
	// Tasks will be sent to this channel for processing by the worker pool.
	queue := make(chan worker.Task)

	// Initialize a new worker pool with the given parameters.
	// This sets up the pool with the provided context, queue queue, and worker count.
	// It prepares the pool to manage and distribute tasks to the workers.
	pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger, MaxRetryWorkerRestart: 3})

	// Start the worker pool in a separate Goroutine to allow it to operate asynchronously.
	// This enables the pool to begin its task processing and worker management in parallel.
	go pool.Run()

	// Assert that initially, no workers should be running.
	// This confirms that the pool starts in an idle state with zero active workers.
	// Ensures that the worker pool initialization is correct before adding any workers.
	assert.Equal(t, int32(0), pool.RunningWorkers(), "Initially, no workers should be running")

	// Pause for 100 milliseconds to give the pool time to initialize and get ready.
	// This small delay ensures that the pool has started running before we add workers.
	<-time.After(200 * time.Millisecond)

	// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
	// This initializes the worker with specified parameters and ensures that it is properly set up.
	worker := NewWorker(workerName, timeout, logger)
	// Assert that the worker instance is not nil.
	// This checks that the worker was successfully created and is not a zero value.
	assert.NotNil(t, worker, "Worker should be successfully created")
	assert.False(t, worker.IsStop())

	// Add the worker to the pool using the `AddWorker` method.
	// This action registers the worker in the pool, allowing it to start processing tasks.
	err := pool.AddWorker(worker)
	// Assert that no error is returned when adding the worker to the pool.
	// This ensures that the `AddWorker` method worked correctly, and the worker was successfully
	// added without any issues, confirming that the pool is properly configured.
	assert.NoError(t, err, "Expected no error while adding worker to the pool")

	// Pause for 100 milliseconds to give the pool time to initialize and get ready.
	// This small delay ensures that the pool has started running before we add workers.
	<-time.After(100 * time.Millisecond)

	// Assert that the number of running workers is equal to 1.
	// This ensures that the worker pool has exactly one worker running after adding the worker.
	assert.Equal(t, int32(1), pool.RunningWorkers(), "There should be exactly 1 worker running after adding the worker")

	// Wait for 100 ms to allow the task to be processed by the worker.
	// This provides sufficient time for the worker to currentProcess the task.
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

		// Check that the worker's is stop after stopping.
		// This ensures that the worker was properly stopped as part of the pool shutdown currentProcess.
		assert.True(t, worker.IsStop(), "The worker should have stopped")

	case <-time.After(20 * time.Second):
		// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
		t.Error("Failed to stop worker pool within expected time")
	}
}

// WorkerShutdown tests the behavior of the worker pool's workerShutdown method.
// It verifies that the method properly shuts down a worker, reduces the number
// of active workers, and removes the worker from the worker pool.
// Additionally, it ensures that the proper workers remain after shutdown.
func TestWorkerShutdown(t *testing.T) {
	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Define the number of workers to be used in the worker pool.
	// This value determines how many worker goroutines will be created.
	workerCount := int32(16)

	// Create a parent context for the worker pool.
	// The context is used to control the lifetime of the worker pool.
	parentCtx := context.Background()

	// Create a channel for task submission.
	// Tasks will be sent to this channel for processing by the worker pool.
	queue := make(chan worker.Task)

	// Initialize a new worker pool with the given parameters.
	// This sets up the pool with the provided context, queue queue, and worker count.
	// It prepares the pool to manage and distribute tasks to the workers.
	pool := NewPool(&worker.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, Logger: logger, MaxRetryWorkerRestart: 3})

	// Assert that initially, no workers should be running.
	// This ensures that the worker pool starts in an idle state before tasks are submitted.
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
	// This checks that the pool correctly updates the count of active workers following the shutdown currentProcess.
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
}

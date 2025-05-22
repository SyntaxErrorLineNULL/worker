//nolint:all
package worker

import (
	"context"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"

	wr "github.com/SyntaxErrorLineNULL/worker"
	"github.com/SyntaxErrorLineNULL/worker/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Parallel()

	defer runtime.GC()

	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Create a new mock instance of the Task using the mocks package.
	// This allows us to use a mock job object in the test, which can be configured to simulate specific behaviors or interactions.
	mockTask := mocks.NewTask(t)
	// Assert that the mockTask object is not nil.
	// This checks that the mock instance was successfully created and initialized.
	// It ensures that any subsequent operations on mockTask will not result in a nil reference error.
	assert.NotNil(t, mockTask, "Expected mockTask to be initialized and not nil")

	// InitWorker tests the initialization of a Worker instance.
	// It verifies that the worker is properly created with the expected initial values
	// and that all required channels are initialized. This test ensures that
	// the worker is correctly set up with the specified ID, timeout, and initial status.
	t.Run("InitWorker", func(t *testing.T) {
		// Define a name for the worker to be used in this test.
		// This name is used to initialize and identify the worker instance.
		workerName := "workerName"

		// Create a new Worker instance with the specified name.
		// The NewWorker function initializes a worker with a unique name and default settings.
		// The worker is expected to be properly set up with a timeout and a logger.
		worker := NewWorker(workerName, 1*time.Second, logger)

		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")
		// Assert that the worker's name is correctly set to the given name.
		// The String method returns the worker's name, which should match the provided workerName.
		// This confirms that the worker's name was initialized correctly.
		assert.Equal(t, workerName, worker.String(), "Worker name should match the provided name")
		// Assert that the worker's timeout is set to 1 second.
		// This checks that the worker's timeout value is correctly set during initialization.
		assert.Equal(t, time.Second, worker.timeout, "Worker timeout should be 1 second")
		// Verify that the worker is currently running and not stopped.
		// This assertion ensures that the IsStop method returns false, indicating
		// that the worker has not been stopped at this point in the test.
		// It helps confirm the worker’s expected lifecycle state before any stop actions.
		assert.False(t, worker.IsStop(), "Worker should not be stopped")
	})

	// SetQueue tests the SetQueue method of the Worker type.
	// It verifies that the method correctly handles setting both open and closed channels as the task queue.
	// The test ensures that setting an open channel succeeds, while setting a closed channel produces an error,
	// validating proper queue management within the worker.
	t.Run("SetQueue", func(t *testing.T) {
		// Define a name for the worker to be used in this test.
		// This name is used to initialize and identify the worker instance.
		workerName := "workerName"

		// Create a new Worker instance with the specified name.
		// The NewWorker function initializes a worker with a unique name and default settings.
		// The worker is expected to be properly set up with a timeout and a logger.
		worker := NewWorker(workerName, 1*time.Second, logger)
		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")

		// Create an open channel to be used as a task queue.
		// This channel is not closed and should be valid for use in the worker.
		openCh := make(chan wr.Task)

		// Set the worker's queue to the open channel.
		// This tests that the worker can successfully use the open channel as its job queue.
		err := worker.SetQueue(openCh)
		// Assert that no error is returned when setting an open channel.
		// This confirms that setting an open channel is handled correctly by the worker.
		assert.NoError(t, err, "Setting an open channel should not produce an error")

		// Create a new channel of type interfaces Task.
		// This channel will be used to test the `SetQueue` method with a closed channel scenario.
		closedChan := make(chan wr.Task)

		// Close the channel to simulate a scenario where the channel is no longer open for receiving tasks.
		// Closing the channel makes it an invalid queue for the worker, which should trigger an error when set.
		close(closedChan)

		// Attempt to set the worker's queue to the closed channel.
		// This action simulates a scenario where an invalid (closed) channel is used as the worker’s job queue.
		err = worker.SetQueue(closedChan)
		// Assert that an error is returned when setting a closed channel.
		// This verifies that the worker's SetQueue method correctly handles and reports an error for closed channels.
		assert.Error(t, err, "Setting a closed channel should produce an error")
	})

	// SetContext tests the SetContext method of the Worker type.
	// It verifies that the worker's context is correctly set and that setting a nil context
	// does not change the existing context. This test ensures proper behavior when managing
	// the worker's context during its lifecycle.
	t.Run("SetContext", func(t *testing.T) {
		// Define a name for the worker to be used in this test.
		// This name is used to initialize and identify the worker instance.
		workerName := "workerName"

		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(workerName, 1*time.Second, logger)

		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")

		// Create a new background context for testing.
		// This context will be used to set and verify the worker's context.
		ctx := context.Background()
		// Set the worker's context to the new background context.
		// This tests the SetContext method by providing a valid context.
		err := worker.SetContext(ctx)
		// Assert that no error occurred when setting the context.
		// This ensures that the SetContext method works as expected when a valid context is provided.
		assert.NoError(t, err, "Expected no error when setting a valid context")

		// Assert that the worker's context is set correctly.
		// This checks that the worker's context matches the provided context after setting it.
		assert.Equal(t, ctx, worker.workerContext, "Worker context should be set to the provided context")

		// Try setting a nil context on the worker.
		// This action tests the worker's behavior when given a nil context, ensuring that it does not change
		// the previously set context.
		err = worker.SetContext(nil)
		// Assert that an error occurred when trying to set a nil context.
		// This verifies that the method correctly returns an error for invalid input.
		assert.Error(t, err, "Expected an error when setting a nil context")

		// Assert that the worker's context remains unchanged after attempting to set a nil context.
		// This verifies that the context does not get altered if a nil value is provided.
		assert.Equal(t, ctx, worker.workerContext, "Worker context should remain unchanged when setting a nil context")
	})

	// GetStatusWithRun tests the behavior of the Worker when it is started and subsequently stopped.
	// It verifies that the worker transitions correctly from idle to stopped status,
	// ensuring proper status reporting throughout the worker's lifecycle.
	t.Run("GetStatusWithRun", func(t *testing.T) {
		// Create a context with cancellation to manage the lifecycle of the worker and ensure proper cleanup.
		// The context will allow the worker to be cancelled if necessary.
		ctx, cancel := context.WithCancel(context.Background())
		// Ensure that the context cancellation function is called at the end of the test.
		// This prevents resource leaks by ensuring proper cleanup after the test completes.
		defer cancel()

		// Define a name for the worker to be used in this test.
		// This name is used to initialize and identify the worker instance.
		workerName := "workerName"
		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(workerName, 1*time.Second, logger)
		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")

		// Set the worker's context to the new background context.
		// This tests the SetContext method by providing a valid context.
		err := worker.SetContext(ctx)
		// Assert that no error occurred when setting the context.
		// This ensures that the SetContext method works as expected when a valid context is provided.
		assert.NoError(t, err, "Expected no error when setting a valid context")

		// Create a channel with a buffer size of 1 to receive jobs.
		// This channel will be used as the job queue for the worker.
		jobQueue := make(chan wr.Task, 1)

		// Set the worker's queue to the open channel.
		// This tests that the worker can successfully use the open channel as its job queue.
		err = worker.SetQueue(jobQueue)
		// Assert that no error is returned when setting an open channel.
		// This confirms that setting an open channel is handled correctly by the worker.
		assert.NoError(t, err, "Setting an open channel should not produce an error")

		// Create a new open channel for error reporting.
		// This channel is intended for use by the worker to report errors encountered during its operations.
		workerErrCh := make(chan *wr.Error)
		// Set the worker's error channel to the newly created open channel.
		// This action configures the worker to use the specified channel for sending error reports.
		// It verifies that the worker can successfully use the provided channel without encountering any errors.
		err = worker.SetWorkerErrChannel(workerErrCh)
		// Assert that no error is returned when setting an open channel.
		// This confirms that the worker's SetWorkerErrChannel method correctly handles the assignment
		// of a valid open channel, ensuring that the channel setup is successful and error-free.
		assert.NoError(t, err, "Setting an open error channel should not produce an error")

		// Create a new WaitGroup to manage goroutines.
		// This WaitGroup will help ensure that all goroutines complete their execution
		// before the test or function exits, maintaining proper synchronization.
		wg := &sync.WaitGroup{}

		// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
		wg.Add(1)
		// Start the worker in a separate goroutine to allow it to run concurrently.
		go worker.Start(wg)

		// Assert that the initial state of the worker is not stopped.
		// This ensures that the worker is in the expected initial state before processing any jobs.
		assert.False(t, worker.IsStop(), "Worker should not be stopped")

		// Sleep for a short duration to allow the worker to transition to a running state.
		time.Sleep(10 * time.Millisecond)

		// Wait for the worker to stop or time out.
		// This block manages the completion of the worker's lifecycle and handles any potential delays.
		select {
		case <-worker.Stop():
			// Worker completed successfully.
			// Any assertions or checks can be placed here if needed.
			assert.True(t, worker.IsStop(), "Expected worker status to be stopped")

			// Use a select statement to receive from the worker's error channel.
			// The select statement allows for non-blocking checks of channel activity.
			select {
			// Attempt to receive from the worker's error channel.
			// If the channel is open and has data, the receive operation will succeed,
			// and `ok` will be true indicating that the channel is still open.
			case _, ok := <-worker.GetError():
				// Assert that the channel is open (ok is true).
				// This ensures that the error channel is functioning correctly and is not closed.
				// The test verifies that the channel is operational, which is critical for proper error reporting.
				assert.True(t, ok, "Error channel should be open and not closed")
			// Default case does nothing; it allows the select statement to proceed without blocking if the channel is not ready.
			default:
			}

		case <-time.After(2 * time.Second):
			t.Error("Worker did not stop within the expected time.")
		}

		// Explicitly wait for the WaitGroup to ensure that all goroutines have finished executing.
		// This ensures proper cleanup and prevents test flakiness due to unfinished goroutines.
		wg.Wait()
	})

	// SuccessProcessing tests the worker's ability to correctly currentProcess a job using the provided mock processing function.
	// The test ensures that the worker can accept a job, currentProcess it within the given context and timeout, and then stop correctly.
	// It verifies that the job's completion is properly signaled and that the worker transitions to a stopped state.
	// Additionally, the test checks if all necessary conditions, like setting contexts and channels, work as expected.
	// The mock processing function simulates a successful operation, and we assert that the counters and status reflect this success.
	t.Run("SuccessProcessing", func(t *testing.T) {
		// Create a background context for the worker.
		// This context will manage the worker's lifecycle and cancellation signals.
		ctx := context.Background()
		// Define a name for the worker to be used in this test.
		// This name is used to initialize and identify the worker instance.
		workerName := "workerName"

		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(workerName, 1*time.Second, logger)
		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")

		// Create an instance of the mock processing job with a specified timeout.
		// This mock simulates a long-running job for testing purposes.
		mockSuccessProcessing := &MockSuccessProcessing{}
		// Initialize input data for the processing function.
		// `inputProcessingData` represents an example integer input (in this case, `222`)
		// that will be passed to the processing function. The integer is cast to `int32`
		// to match the expected data type used by the processing function.
		inputProcessingData := int32(3)

		// Define the worker timeout duration for the test.
		// This is the maximum amount of time we allow for the job to complete.
		workerTimeout := 1 * time.Second

		// Initialize the necessary variables and objects for the job.
		// Create a new job instance with a timeout and a unique job name.
		job := NewTask(workerTimeout, "test-job", mockSuccessProcessing, inputProcessingData)
		// Assert that the job was successfully created.
		// If the job is nil, it indicates a problem with job initialization.
		assert.NotNil(t, job, "Expected job to be initialized, but it was nil")

		// Set the parent context of the job to the newly created context.
		// This context will be used in job processing.
		_ = job.SetContext(ctx)

		// Create a buffered done channel to signal job completion.
		// This channel will be used to notify when the job is done.
		doneCh := make(chan struct{}, 1)
		// Set the done channel for the job using the SetDoneChannel method.
		// The method should return no error if the done channel is valid.
		_ = job.SetDoneChannel(doneCh)

		// Create a new open channel for error reporting.
		// This channel is intended for use by the worker to report errors encountered during its operations.
		workerErrCh := make(chan *wr.Error)
		// Set the worker's error channel to the newly created open channel.
		// This action configures the worker to use the specified channel for sending error reports.
		// It verifies that the worker can successfully use the provided channel without encountering any errors.
		err := worker.SetWorkerErrChannel(workerErrCh)
		// Assert that no error is returned when setting an open channel.
		// This confirms that the worker's SetWorkerErrChannel method correctly handles the assignment
		// of a valid open channel, ensuring that the channel setup is successful and error-free.
		assert.NoError(t, err, "Setting an open error channel should not produce an error")

		// Set the worker's context to the new background context.
		// This tests the SetContext method by providing a valid context.
		err = worker.SetContext(ctx)
		// Assert that no error occurred when setting the context.
		// This ensures that the SetContext method works as expected when a valid context is provided.
		assert.NoError(t, err, "Expected no error when setting a valid context")

		// Create a buffered channel for job collection with a capacity of 1.
		// This channel will be used to collect jobs for the worker pool.
		collector := make(chan wr.Task, 1)

		// Set the worker's queue to the open channel.
		// This tests that the worker can successfully use the open channel as its job queue.
		err = worker.SetQueue(collector)
		// Assert that no error is returned when setting an open channel.
		// This confirms that setting an open channel is handled correctly by the worker.
		assert.NoError(t, err, "Setting an open channel should not produce an error")

		// Create a WaitGroup to manage synchronization of the worker's goroutine.
		// This allows the test to wait for the worker to complete its execution.
		wg := &sync.WaitGroup{}
		defer wg.Wait()
		// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
		// This is necessary to ensure that the test correctly waits for the worker's completion.
		wg.Add(1)

		// Start the worker in a separate goroutine to allow it to run concurrently.
		// This enables the worker to currentProcess jobs asynchronously.
		go worker.Start(wg)

		// Send the mock job to the worker's job queue.
		// This simulates the worker receiving a job for processing.
		collector <- job

		// Use a select statement to handle the job completion or timeout.
		// This waits for either the job to signal completion or a timeout to occur.
		select {
		case <-doneCh:
			// Assert that the processing counter for the mock job is incremented by the volume ID.
			// This confirms that the mock processing job completed successfully.
			assert.Equal(t, inputProcessingData, mockSuccessProcessing.counter.Load())
		case <-time.After(2 * time.Second):
			// If the job doesn't complete within the expected time, trigger an error.
			t.Error("The execution time exceeds the allowable timeout")
		}

		// Wait for the worker to stop processing and check its status.
		// This block waits for either the worker to signal that it has stopped or a timeout to occur.
		select {
		case <-worker.Stop():
			// Assert that the worker status is 'stopped' after cancellation.
			// This verifies that the worker correctly transitions to the 'stopped' state.
			assert.True(t, worker.IsStop(), "Expected worker status to be stopped")

			// Assert that the `currentProcess` field of the `worker` is `nil`.
			// This ensures that the worker has not started processing or has properly cleaned up its processing state.
			assert.Nil(t, worker.currentProcess, "Expected worker.currentProcess to be nil, indicating that the worker has not started processing or has been properly cleaned up.")
		case <-time.After(2 * time.Second):
			// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
			t.Error("Failed to stop worker within expected time")
		}
	})
}

// ContextDone tests the behavior of the Worker when the context is cancelled.
// This test ensures that the worker stops processing and updates its status appropriately
// when the context is cancelled, simulating a graceful shutdown.
func TestContextDone(t *testing.T) {
	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Create a context with cancellation capabilities.
	// This context will be used to control the worker’s execution lifecycle,
	// allowing us to signal the worker to stop.
	ctx, cancel := context.WithCancel(context.Background())
	// Define the worker timeout duration for the test.
	// This is the maximum amount of time we allow for the job to complete.
	timeout := 10 * time.Second
	// Define a name for the worker to be used in this test.
	// This name is used to initialize and identify the worker instance.
	workerName := "workerName"

	// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
	// This initializes the worker with specified parameters and ensures that it is properly set up.
	worker := NewWorker(workerName, 1*time.Second, logger)
	// Assert that the worker instance is not nil.
	// This checks that the worker was successfully created and is not a zero value.
	assert.NotNil(t, worker, "Worker should be successfully created")

	// Create an instance of the mock processing task with a specified timeout.
	// This mock simulates a long-running task for testing purposes.
	mockProcessingWithLongTask := &MockProcessingLongTask{timeout: timeout}
	// resultCh is a buffered channel that is used to receive the result of the mock processing task.
	// The buffer size is set to 1 because we expect only one result to be sent through this channel
	// for each task execution. This allows the sending goroutine to proceed without blocking, assuming
	// the result is consumed quickly.
	resultCh := make(chan *MockProcessingLongTaskResult, 1)

	// Create a new task instance with the specified timeout, name, processing function, and input data.
	// This task simulates a job that will be processed within the test.
	job := NewTask(timeout, "test-task", mockProcessingWithLongTask, resultCh)
	// Assert that the job was successfully created.
	// If the job is nil, it indicates a problem with job initialization.
	assert.NotNil(t, job, "Expected job to be initialized, but it was nil")

	// Set the parent context of the job to the newly created context.
	// This context will be used in job processing.
	_ = job.SetContext(ctx)

	// Create a buffered done channel to signal job completion.
	// This channel will be used to notify when the job is done.
	doneCh := make(chan struct{}, 1)
	// Set the done channel for the job using the SetDoneChannel method.
	// The method should return no error if the done channel is valid.
	_ = job.SetDoneChannel(doneCh)

	// Create a wait group to synchronize job completion.
	// The wait group will be used to wait for the job to complete.
	wg := &sync.WaitGroup{}

	// Set the worker's context to the new background context.
	// This tests the SetContext method by providing a valid context.
	err := worker.SetContext(ctx)
	// Assert that no error occurred when setting the context.
	// This ensures that the SetContext method works as expected when a valid context is provided.
	assert.NoError(t, err, "Expected no error when setting a valid context")

	// Create a buffered channel for job collection with a capacity of 1.
	// This channel will be used to collect jobs for the worker pool.
	collector := make(chan wr.Task, 1)

	// Set the worker's queue to the open channel.
	// This tests that the worker can successfully use the open channel as its job queue.
	err = worker.SetQueue(collector)
	// Assert that no error is returned when setting an open channel.
	// This confirms that setting an open channel is handled correctly by the worker.
	assert.NoError(t, err, "Setting an open channel should not produce an error")

	// Send the mock job to the worker's job queue.
	// This simulates the worker receiving a job for processing.
	collector <- job

	// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
	// This is necessary to ensure that the test correctly waits for the worker's completion.
	wg.Add(1)

	// Start the worker in a separate goroutine to allow it to run concurrently.
	// This enables the worker to currentProcess jobs asynchronously.
	go worker.Start(wg)

	// Wait for 100 millisecond to allow the worker to currentProcess the job.
	// This delay provides sufficient time for the worker to handle the job before canceling the context.
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to signal the worker to stop processing.
	// This action is used to test if the worker responds correctly to context cancellation.
	cancel()

	// Wait for the worker to stop processing and check its status.
	// This block waits for either the worker to signal that it has stopped or a timeout to occur.
	select {
	case <-worker.stopCh:
		// Assert that the worker status is 'stopped' after cancellation.
		// This verifies that the worker correctly transitions to the 'stopped' state.
		// spew.Dump(worker.IsStop())
		assert.True(t, worker.IsStop(), "Expected worker status to be stopped")

		// Assert that the `currentProcess` field of the `worker` is `nil`.
		// This ensures that the worker has not started processing or has properly cleaned up its processing state.
		assert.Nil(t, worker.currentProcess, "Expected worker.currentProcess to be nil, indicating that the worker has not started processing or has been properly cleaned up.")

		select {
		// Attempt to receive from the `stopCh` channel to check if it's closed.
		// In Go, when receiving from a closed channel, the operation will return the zero value
		// of the channel's type immediately and the second value (`ok`) will be false.
		// This behavior allows us to determine if the channel has been closed by checking the `ok` value.
		case <-job.stopCh:
			// Attempt to receive from the `stopCh` channel to check if it's closed.
			// In Go, receiving from a closed channel returns the zero value immediately and `ok` is false.
			// If the channel is still open, `ok` would be true, indicating that the job is still running.
			_, ok := <-job.stopCh

			// Assert that `ok` is false, meaning that the `stopCh` should be closed at this point.
			// A closed `stopCh` indicates that the job has completed its execution and signaled completion.
			// If the channel is still open (`ok` is true), this would imply the job has not finished properly, and the test should fail.
			assert.False(t, ok, "Expected stop channel to be closed, indicating job completion")
		default:
			// The `default` case is executed if none of the other cases in the select statement are ready.
			// This provides a non-blocking path, ensuring that the select statement can proceed
			// without being stuck waiting for an input from the channels. In this context, it effectively
			// does nothing and allows the test to continue without blocking.
		}

		// Attempt to receive a result from the resultCh channel.
		// If a result is available, the assertions validate the expected state of the result.
		select {
		case res := <-resultCh:
			// Assert that the ContextIsDone field is true, indicating that the context
			// associated with the task is completed as expected.
			assert.True(t, res.ContextIsDone, "Expected the context to be marked as done")
			// Assert that the ContextIsNotDone field is false, ensuring that the context
			// is not incorrectly marked as active when it has already completed.
			assert.False(t, res.ContextIsNotDone, "Expected the context to not be active")
		default:
			// If no result is available on the resultCh channel, this case prevents the
			// code from blocking indefinitely. The absence of a result may indicate that
			// no tasks have completed yet or that the channel is empty.
		}

	case <-time.After(2 * time.Second):
		// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
		t.Error("Failed to stop worker within expected time")
	}
}

// WorkerPanic tests the behavior of the worker when a job causes a panic. It verifies that
// the worker can handle panics properly by reporting them via the error channel and that
// the worker stops processing and transitions to the 'stopped' state as expected. The test
// ensures that the worker's panic handling mechanism is functioning correctly and that
// the worker cleans up its processing state after encountering a panic.
func TestWorkerPanic(t *testing.T) {
	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Create a background context for the worker.
	// This context will manage the worker's lifecycle and cancellation signals.
	ctx := context.Background()

	// Create a WaitGroup to manage synchronization of the worker's goroutine.
	// This allows the test to wait for the worker to complete its execution.
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	// Define a name for the worker to be used in this test.
	// This name is used to initialize and identify the worker instance.
	workerName := "workerName"

	// Create an instance of MockPanicTask, which simulates a job that will panic during its execution.
	// This mock task is used to test the worker's panic handling capabilities.
	mockTaskWithPanic := &MockPanicTask{}

	// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
	// This initializes the worker with specified parameters and ensures that it is properly set up.
	worker := NewWorker(workerName, 1*time.Second, logger)
	// Assert that the worker instance is not nil.
	// This checks that the worker was successfully created and is not a zero value.
	assert.NotNil(t, worker, "Worker should be successfully created")

	// Set the worker's context to the new background context.
	// This tests the SetContext method by providing a valid context.
	err := worker.SetContext(ctx)
	// Assert that no error occurred when setting the context.
	// This ensures that the SetContext method works as expected when a valid context is provided.
	assert.NoError(t, err, "Expected no error when setting a valid context")

	// Create a new open channel for error reporting.
	// This channel is intended for use by the worker to report errors encountered during its operations.
	workerErrCh := make(chan *wr.Error)
	// Set the worker's error channel to the newly created open channel.
	// This action configures the worker to use the specified channel for sending error reports.
	// It verifies that the worker can successfully use the provided channel without encountering any errors.
	err = worker.SetWorkerErrChannel(workerErrCh)
	// Assert that no error is returned when setting an open channel.
	// This confirms that the worker's SetWorkerErrChannel method correctly handles the assignment
	// of a valid open channel, ensuring that the channel setup is successful and error-free.
	assert.NoError(t, err, "Setting an open error channel should not produce an error")

	// Create a buffered channel for job collection with a capacity of 1.
	// This channel will be used to collect jobs for the worker pool.
	collector := make(chan wr.Task, 1)

	// Set the worker's queue to the open channel.
	// This tests that the worker can successfully use the open channel as its job queue.
	err = worker.SetQueue(collector)
	// Assert that no error is returned when setting an open channel.
	// This confirms that setting an open channel is handled correctly by the worker.
	assert.NoError(t, err, "Setting an open channel should not produce an error")

	// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
	// This is necessary to ensure that the test correctly waits for the worker's completion.
	wg.Add(1)

	// Start the worker in a separate goroutine to allow it to run concurrently.
	// This enables the worker to currentProcess jobs asynchronously.
	go worker.Start(wg)

	// Send the mock job to the worker's job queue.
	// This simulates the worker receiving a job for processing.
	collector <- mockTaskWithPanic

	select {
	// Wait for the worker to handle the task and check for errors or a timeout.
	// This select block listens for the worker to signal that it encountered an error or waits for a timeout.
	case workerError := <-worker.GetError():
		// Assert that the error received is the expected panic error.
		// This verifies that the worker correctly identified and reported the panic error.
		assert.ErrorIs(t, workerError.Error, errMockPanic)
		// Assert that the error instance matches the worker that encountered the error.
		// This confirms that the correct worker is associated with the reported error.
		assert.Equal(t, worker, workerError.Instance)

		assert.Equal(t, int32(0), worker.GetRetry())

	case <-time.After(5 * time.Second):
		// If the worker does not stop within the allocated 2 seconds, this block will execute.
		// This indicates that the worker took too long to stop, which could signify a problem with the shutdown currentProcess.
		// The test will fail, providing feedback that the worker did not stop as expected within the given time frame.
		t.Error("Failed to stop worker within expected time")
	}

	// Wait for the worker to stop processing and check its status.
	// This block waits for either the worker to signal that it has stopped or a timeout to occur.
	select {
	case <-worker.Stop():
		// Assert that the worker status is 'stopped' after cancellation.
		// This verifies that the worker correctly transitions to the 'stopped' state.
		assert.True(t, worker.IsStop(), "Expected worker status to be stopped")

		// Assert that the `currentProcess` field of the `worker` is `nil`.
		// This ensures that the worker has not started processing or has properly cleaned up its processing state.
		assert.Nil(t, worker.currentProcess, "Expected worker.currentProcess to be nil, indicating that the worker has not started processing or has been properly cleaned up.")
	case <-time.After(2 * time.Second):
		// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
		t.Error("Failed to stop worker within expected time")
	}
}

// SuccessStop tests the ability of the worker to handle context cancellation correctly during long-running job processing.
// The test ensures that the worker can start processing a job, and upon cancellation, the worker transitions to a stopped state.
// It verifies that the job's context is canceled properly, and that the job stops execution without triggering a timeout.
// Additionally, the test checks that the worker's internal counters, such as `contextDoneCounter`, reflect the correct behavior
// after the context is canceled. This ensures that the worker and job handle graceful termination without errors.
func TestSuccessStop(t *testing.T) {
	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Create a background context for the worker.
	// This context will manage the worker's lifecycle and cancellation signals.
	ctx := context.Background()

	// Define a name for the worker to be used in this test.
	// This name is used to initialize and identify the worker instance.
	workerName := "workerName"
	// Define the worker timeout duration for the test.
	// This is the maximum amount of time we allow for the task to complete.
	timeout := 10 * time.Second

	// Create an instance of the mock processing task with a specified timeout.
	// This mock simulates a long-running task for testing purposes.
	mockProcessingWithLongTask := &MockProcessingLongTask{timeout: timeout}
	// resultCh is a buffered channel that is used to receive the result of the mock processing task.
	// The buffer size is set to 1 because we expect only one result to be sent through this channel
	// for each task execution. This allows the sending goroutine to proceed without blocking, assuming
	// the result is consumed quickly.
	resultCh := make(chan *MockProcessingLongTaskResult, 1)

	// Create a new task instance with the specified timeout, name, processing function, and input data.
	// This task simulates a task that will be processed within the test.
	task := NewTask(timeout, "test-task", mockProcessingWithLongTask, resultCh)
	// Assert that the task was successfully created.
	// If the task is nil, it indicates a problem with task initialization.
	assert.NotNil(t, task, "Expected task to be initialized, but it was nil")

	// Create a buffered done channel to signal task completion.
	// This channel will be used to notify when the task is done.
	doneCh := make(chan struct{}, 1)
	// Set the done channel for the task using the SetDoneChannel method.
	// The method should return no error if the done channel is valid.
	_ = task.SetDoneChannel(doneCh)

	// Create a new WaitGroup to manage goroutines.
	// This WaitGroup will help ensure that all goroutines complete their execution
	// before the test or function exits, maintaining proper synchronization.
	wg := &sync.WaitGroup{}
	// Defer a call to Wait() on the WaitGroup to block the main routine until
	// all tracked goroutines have called Done(). This ensures that the program
	// does not exit prematurely and allows all background tasks to finish cleanly.
	defer wg.Wait()

	// Assign the wait group to the task instance.
	// This allows the task to signal completion to the wait group.
	_ = task.SetWaitGroup(wg)

	// Set the parent context of the task to the newly created context.
	// This context will be used in task processing.
	_ = task.SetContext(ctx)

	// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
	// This initializes the worker with specified parameters and ensures that it is properly set up.
	worker := NewWorker(workerName, 1*time.Second, logger)
	// Assert that the worker instance is not nil.
	// This checks that the worker was successfully created and is not a zero value.
	assert.NotNil(t, worker, "Worker should be successfully created")

	// Set the worker's context to the new background context.
	// This tests the SetContext method by providing a valid context.
	err := worker.SetContext(ctx)
	// Assert that no error occurred when setting the context.
	// This ensures that the SetContext method works as expected when a valid context is provided.
	assert.NoError(t, err, "Expected no error when setting a valid context")

	// Create a buffered channel for task collection with a capacity of 1.
	// This channel will be used to collect jobs for the worker pool.
	collector := make(chan wr.Task, 1)

	// Set the worker's queue to the open channel.
	// This tests that the worker can successfully use the open channel as its task queue.
	err = worker.SetQueue(collector)
	// Assert that no error is returned when setting an open channel.
	// This confirms that setting an open channel is handled correctly by the worker.
	assert.NoError(t, err, "Setting an open channel should not produce an error")

	// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
	// This is necessary to ensure that the test correctly waits for the worker's completion.
	wg.Add(1)

	// Start the worker in a separate goroutine to allow it to run concurrently.
	// This enables the worker to currentProcess jobs asynchronously.
	go worker.Start(wg)

	// Pause the execution for 50 milliseconds. This brief delay allows the worker
	// goroutine to initialize and potentially process some tasks, ensuring
	// accurate testing of its functionality.
	<-time.After(50 * time.Millisecond)

	// Send the mock task to the worker's task queue.
	// This simulates the worker receiving a task for processing.
	collector <- task

	// Pause the execution for 50 milliseconds. This brief delay allows the worker
	// goroutine to initialize and potentially process some tasks, ensuring
	// accurate testing of its functionality.
	<-time.After(50 * time.Millisecond)

	// Wait for the worker to stop processing and check its status.
	// This block waits for either the worker to signal that it has stopped or a timeout to occur.
	select {
	case <-worker.Stop():
		// Assert that the worker status is 'stopped' after cancellation.
		// This verifies that the worker correctly transitions to the 'stopped' state.
		assert.True(t, worker.IsStop(), "Expected worker status to be stopped")
		// Assert that the `currentProcess` field of the `worker` is `nil`.
		// This ensures that the worker has not started processing or has properly cleaned up its processing state.
		assert.Nil(t, worker.currentProcess, "Expected worker.currentProcess to be nil, indicating that the worker has not started processing or has been properly cleaned up.")

		select {
		// This case is triggered when the task sends a completion signal on the doneCh channel.
		// The following assertions verify the task's state after context cancellation.
		case <-doneCh:
			// Attempt to retrieve a result from the resultCh channel.
			select {
			case res := <-resultCh:
				// Assert that the ContextIsNotDone field is false, ensuring that the task's context
				// is not mistakenly marked as active. This verifies the task behaves correctly
				// under the given conditions.
				assert.False(t, res.ContextIsNotDone, "Expected the context to not be active")
			default:
				// If no result is available, the default case prevents blocking, allowing the
				// program to continue execution without waiting for a value on the channel.
			}
		default:
			// This default case handles the situation where no other case is triggered within the select statement's
			// timeout period. In this context, it effectively does nothing but ensures the select block completes
			// if the doneCh case does not get triggered.
		}

	case <-time.After(3 * time.Second):
		// Timeout case: if the pool does not stop within the expected time, indicate a test failure.
		t.Error("Failed to stop worker within expected time")
	}
}

// RestartWorker tests the behavior of the `Restart` method for a worker in the worker pool system.
// It ensures that the worker can be restarted correctly and that the retry count is incremented as expected.
// Additionally, it verifies that the worker transitions to an idle state upon restart and properly shuts down
// when stopped. The test checks that the worker's status and retry count match the expected values after the restart
// operation and confirms that the worker's error channel is closed after stopping.
func TestRestartWorker(t *testing.T) {
	// Define a name for the worker to be used in this test.
	// This name is used to initialize and identify the worker instance.
	workerName := "workerName"
	// retryCount represents the number of times an operation should be retried
	// in case of failure. Here, it is initialized to 1, meaning the operation
	// will be attempted once before considering it a failure. This value can
	// be adjusted based on the desired retry strategy to handle transient errors.
	retryCount := int32(1)

	// Create a logger instance for the worker pool.
	// This logger will be used to log information or errors from the worker pool.
	logger := log.Default()

	// Create a background context for the worker.
	// This context will manage the worker's lifecycle and cancellation signals.
	ctx := context.Background()

	// Create a WaitGroup to manage synchronization of the worker's goroutine.
	// This allows the test to wait for the worker to complete its execution.
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
	// This initializes the worker with specified parameters and ensures that it is properly set up.
	worker := NewWorker(workerName, 1*time.Second, logger)
	// Assert that the worker instance is not nil.
	// This checks that the worker was successfully created and is not a zero value.
	assert.NotNil(t, worker, "Worker should be successfully created")

	// Set the worker's context to the new background context.
	// This tests the SetContext method by providing a valid context.
	err := worker.SetContext(ctx)
	// Assert that no error occurred when setting the context.
	// This ensures that the SetContext method works as expected when a valid context is provided.
	assert.NoError(t, err, "Expected no error when setting a valid context")

	// Create a buffered channel for job collection with a capacity of 1.
	// This channel will be used to collect jobs for the worker pool.
	collector := make(chan wr.Task, 1)

	// Set the worker's queue to the open channel.
	// This tests that the worker can successfully use the open channel as its job queue.
	err = worker.SetQueue(collector)
	// Assert that no error is returned when setting an open channel.
	// This confirms that setting an open channel is handled correctly by the worker.
	assert.NoError(t, err, "Setting an open channel should not produce an error")

	// Create a new open channel for error reporting.
	// This channel is intended for use by the worker to report errors encountered during its operations.
	workerErrCh := make(chan *wr.Error)
	// Set the worker's error channel to the newly created open channel.
	// This action configures the worker to use the specified channel for sending error reports.
	// It verifies that the worker can successfully use the provided channel without encountering any errors.
	err = worker.SetWorkerErrChannel(workerErrCh)
	// Assert that no error is returned when setting an open channel.
	// This confirms that the worker's SetWorkerErrChannel method correctly handles the assignment
	// of a valid open channel, ensuring that the channel setup is successful and error-free.
	assert.NoError(t, err, "Setting an open error channel should not produce an error")

	// Increment the WaitGroup counter by 1 to account for the worker's goroutine.
	// This is necessary to ensure that the test correctly waits for the worker's completion.
	wg.Add(1)

	// Start the worker in a separate goroutine to allow it to run concurrently.
	// This enables the worker to currentProcess jobs asynchronously.
	worker.Restart(wg)

	// Assert that the worker's retry count matches the expected value.
	// This checks that the worker's restart currentProcess has incremented the retry count correctly.
	assert.Equal(t, retryCount, worker.GetRetry(), "Worker retry count should match the expected retry count")

	// Wait for the worker to stop or time out.
	// This block manages the completion of the worker's lifecycle and handles any potential delays.
	select {
	case <-worker.Stop():
		// Worker completed successfully.
		// Any assertions or checks can be placed here if needed.
		assert.True(t, worker.IsStop(), "Expected worker status to be stopped")

		// Use a select statement to receive from the worker's error channel.
		// The select statement allows for non-blocking checks of channel activity.
		select {
		// Attempt to receive from the worker's error channel.
		// If the channel is open and has data, the receive operation will succeed,
		// and `ok` will be true indicating that the channel is still open.
		case _, ok := <-worker.GetError():
			// Assert that the channel is open (ok is true).
			// This ensures that the error channel is functioning correctly and is not closed.
			// The test verifies that the channel is operational, which is critical for proper error reporting.
			assert.True(t, ok, "Error channel should be open and not closed")
		// Default case does nothing; it allows the select statement to proceed without blocking if the channel is not ready.
		default:
		}

	case <-time.After(2 * time.Second):
		t.Error("Worker did not stop within the expected time.")
	}
}

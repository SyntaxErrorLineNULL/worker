package worker

import (
	"context"
	"sync"
	"testing"
	"time"
	"worker/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTask(t *testing.T) {
	t.Parallel()

	// Create a new mock instance of the Processing interface using the mocks package.
	// This mock object simulates the behavior of a Processing interface, allowing you to test how your code interacts with it.
	mockProcessing := mocks.NewProcessing(t)
	// Assert that the mockProcessing object is not nil.
	// This verifies that the mock instance was successfully created and initialized.
	// It ensures that the mock object is properly set up for use in the test, avoiding issues related to nil references.
	assert.NotNil(t, mockProcessing, "Expected mockProcessing to be initialized and not nil")

	// SetWaitGroup tests the SetWaitGroup method of the Task struct.
	// It verifies that the method correctly assigns a provided WaitGroup to the Task,
	// and ensures that no errors occur when a valid WaitGroup is passed.
	t.Run("SetWaitGroup", func(t *testing.T) {
		// Initialize a new WaitGroup.
		var wg sync.WaitGroup
		// Create a new instance of task.
		job := &Task{}

		// Set the wait group for the job using the SetWaitGroup method.
		err := job.SetWaitGroup(&wg)
		// Assert that no error was returned from SetWaitGroup.
		// This check confirms that the method handled the valid WaitGroup correctly.
		assert.NoError(t, err, "expected no error when setting a valid WaitGroup")

		// Assert that the wait group was set correctly in the task.
		// The expected value is the address of the initialized WaitGroup.
		// The actual value is the task's wg field.
		// The message "expected the task's wait group to be set correctly" is displayed if the assertion fails.
		assert.Equal(t, &wg, job.wg, "expected the job's wait group to be set correctly")
	})

	// SetEmptyWaitGroup tests the SetWaitGroup method of the Task struct
	// when attempting to set a nil WaitGroup. This test ensures that the method
	// correctly handles the case where a nil WaitGroup is provided, returning
	// an appropriate error message.
	t.Run("SetEmptyWaitGroup", func(t *testing.T) {
		// Create a new instance of Task.
		job := &Task{}

		// Attempt to set a nil WaitGroup for the Task using the SetWaitGroup method.
		// The method is expected to return an error in this case.
		err := job.SetWaitGroup(nil)

		// Assert that an error was returned from SetWaitGroup.
		// This check confirms that the method properly handles the nil WaitGroup.
		assert.Error(t, err, "expected error when setting a nil WaitGroup")

		// Assert that the error message is as expected.
		// The error should indicate that the WaitGroup cannot be nil.
		assert.Equal(t, "WaitGroup cannot be nil", err.Error(), "expected a specific error message for nil WaitGroup")
	})

	// SetDoneChannel tests the SetDoneChannel method of the Task struct.
	// It ensures that the method correctly assigns a provided done channel to the Task,
	// and verifies that no errors occur during this process.
	t.Run("SetDoneChannel", func(t *testing.T) {
		// Create a done channel.
		doneCh := make(chan struct{})
		// Create a new instance of task.
		task := &Task{}

		// Set the done channel for the task using the SetDoneChannel method.
		// The method should return no error if the done channel is valid.
		err := task.SetDoneChannel(doneCh)
		// Assert that no error was returned from SetDoneChannel.
		// This check confirms that the method handled the done channel correctly.
		assert.NoError(t, err, "expected no error when setting a valid done channel")

		// Assert that the done channel was set correctly in the task.
		// The expected value is the done channel cast to a send-only channel.
		// The actual value is the task's doneCh field.
		assert.Equal(t, (chan<- struct{})(doneCh), task.doneCh, "expected the task's done channel to be set correctly")
	})

	// SetEmptyChannel tests the SetDoneChannel method of the Task struct
	// when attempting to set a nil channel. This test ensures that the method
	// correctly handles the case where a nil channel is provided, returning
	// an appropriate error.
	t.Run("SetEmptyChannel", func(t *testing.T) {
		// Create a new instance of Task.
		task := &Task{}

		// Attempt to set a nil channel for the Task using the SetDoneChannel method.
		// The method is expected to return an error in this case.
		err := task.SetDoneChannel(nil)

		// Assert that an error was returned from SetDoneChannel.
		// This check confirms that the method properly handles the nil channel.
		assert.Error(t, err, "expected error when setting a nil done channel")
	})

	// SetCloseChannel tests the SetDoneChannel method of the Task struct
	// when a closed channel is provided. This test ensures that the method
	// correctly identifies and handles the scenario where a closed channel
	// is passed, and returns an appropriate error.
	t.Run("SetCloseChannel", func(t *testing.T) {
		// Create a done channel using make. This will be used to test the behavior
		// of the SetDoneChannel method when a closed channel is passed.
		doneCh := make(chan struct{})

		// Create a new instance of Task. This represents the object whose
		// SetDoneChannel method will be tested.
		task := &Task{}

		// Close the done channel to simulate an invalid state where the channel
		// is already closed. This is to test how SetDoneChannel handles such cases.
		close(doneCh)

		// Attempt to set the closed channel for the Task using the SetDoneChannel method.
		// The method should identify that the channel is closed and handle it appropriately.
		// The expected outcome is that an error is returned since the channel should not be set
		// if it is closed.
		err := task.SetDoneChannel(doneCh)

		// Assert that an error is returned from SetDoneChannel. This assertion checks
		// that the method correctly identifies and rejects the closed channel.
		// The test will fail if no error is returned, indicating that the method did
		// not handle the closed channel as expected.
		assert.Error(t, err, "expected error when setting a closed done channel")

		// Optionally, you could add an assertion to check the specific
		// error message returned to ensure it matches expected error messages.
		// This is useful if you want to validate the exact reason for the failure.
		assert.Equal(t, "cannot set a closed channel", err.Error(), "unexpected error message")
	})

	// SetContext tests the SetContext method of the Task struct to ensure
	// it correctly sets the internal context of the Task instance. This method
	// is expected to update the Task's context to the provided context value and
	// handle different scenarios such as valid and empty contexts.
	t.Run("SetContext", func(t *testing.T) {
		// Create a new instance of Task. This is the object on which the
		// SetContext method will be tested.
		task := &Task{}

		// Create a background context using context.Background().
		// This context will be used as the input to the SetContext method.
		ctx := context.Background()

		// Call the SetContext method on the Task instance, passing the created context.
		// This method should set the internal context of the Task to the provided context.
		err := task.SetContext(ctx)

		// Assert that no error is returned from the SetContext method. This checks
		// that the method executed successfully without encountering any issues
		// and confirms that it handles the provided context correctly.
		assert.NoError(t, err, "SetContext should not return an error")

		// Assert that the Task's internal context (parentCtx) is correctly set to
		// the context that was passed to the SetContext method. This ensures that
		// the method has updated the internal state of the Task instance as expected.
		// The expected value is the context passed to SetContext, and the actual
		// value is the Task's parentCtx field.
		assert.Equal(t, ctx, task.parentCtx, "expected Task's context to be set correctly")
	})

	// SetEmptyContext tests the SetContext method of the Task struct to ensure
	// it correctly handles the case when a nil context is provided. This test
	// verifies that the method returns an appropriate error when attempting to
	// set the Task's context to nil, which is an invalid operation.
	t.Run("SetEmptyContext", func(t *testing.T) {
		// Create a new instance of Task. This instance will be used to test the
		// SetContext method. Initially, the Task's internal context should be
		// uninitialized or set to its zero value.
		task := &Task{}

		// Call the SetContext method on the Task instance, passing a nil context.
		// The SetContext method is expected to handle the nil context case
		// and return an error because setting a nil context is not valid.
		err := task.SetContext(nil)

		// Assert that the SetContext method returns an error when given a nil context.
		// This checks that the method correctly identifies and handles the invalid input.
		// The assertion will fail if no error is returned or if the error does not match
		// the expected error message.
		assert.Error(t, err, "SetContext should return an error when given a nil context")

		// Assert that the error returned from SetContext matches the expected error message.
		// The expected message is "context cannot be nil", which indicates that the method
		// correctly identifies the issue with a nil context. This ensures that the error
		// handling is properly implemented and that the method provides a clear and
		// informative error message for invalid inputs.
		assert.Equal(t, "context cannot be nil", err.Error(), "error message should be 'context cannot be nil'")
	})

	// GetName tests the String method of the Task struct to ensure that it
	// correctly returns the name of the task. This test verifies that the
	// String method accurately reflects the task's name as stored in the
	// Task struct.
	t.Run("GetName", func(t *testing.T) {
		// Define the expected name that will be assigned to the Task instance.
		name := "test-name"

		// Create a new Task instance and initialize it with the specified name.
		// The Task struct should have a 'name' field that stores this value.
		task := &Task{name: name}

		// Assert that the String method of the Task instance returns the correct name.
		// The String method is expected to return the name of the task as a string.
		// This assertion will fail if the returned value does not match the expected name.
		assert.Equal(t, name, task.String(), "String() should return the task's name")
	})

	// TaskCompletesWithoutTimeout tests the Run method of the Job struct to ensure that a job
	// completes without timing out. It verifies that the job correctly executes its task
	// within the allowed time and signals completion through the done channel.
	t.Run("TaskCompletesWithoutTimeout", func(t *testing.T) {
		// Define the worker timeout duration for the test.
		// This is the maximum amount of time we allow for the task to complete.
		timeout := 1 * time.Second

		// Initialize input data for the processing function.
		// `inputProcessingData` represents an example integer input (in this case, `222`)
		// that will be passed to the processing function. The integer is cast to `int32`
		// to match the expected data type used by the processing function.
		inputProcessingData := int32(222)

		// Define the error handler input as a string.
		// `inputErrorHandler` is set to "error handler" and is intended to be used
		// as the input for the error handling function in the task.
		// This string will simulate or represent an identifier or message to be processed by the error handler.
		inputErrorHandler := "error handler"

		// Create a new task instance with the specified timeout, name, processing function, and input data.
		// This task simulates a job that will be processed within the test.
		task := NewTask(timeout, "test-task", mockProcessing, inputProcessingData, inputErrorHandler)
		// Assert that the task was successfully created.
		// If the task is nil, it indicates a problem with task initialization.
		assert.NotNil(t, task, "Expected task to be initialized, but it was nil")

		// Create a buffered done channel to signal job completion.
		// This channel will be used to notify when the job is done.
		doneCh := make(chan struct{}, 1)
		// Set the done channel for the task using the SetDoneChannel method.
		// The method should return no error if the done channel is valid.
		_ = task.SetDoneChannel(doneCh)

		// Create a wait group to synchronize job completion.
		// The wait group will be used to wait for the task to complete.
		wg := &sync.WaitGroup{}
		// Assign the wait group to the job instance.
		// This allows the task to signal completion to the wait group.
		_ = task.SetWaitGroup(wg)

		// Create a context without a timeout to use for the task.
		// The context is used for task processing and cancellation.
		ctx := context.Background()
		// Set the parent context of the task to the newly created context.
		// This context will be used in task processing.
		_ = task.SetContext(ctx)

		// Mock the Processing method to return true when called with the context and task.
		// This simulates the job processing method returning a successful result.
		mockProcessing.On("Processing", mock.Anything, inputProcessingData).Return(true)

		// Add to the wait group to track job execution.
		// This ensures the wait group waits for the job to complete.
		wg.Add(1)

		// Start the job in a separate Goroutine to allow asynchronous execution.
		// This allows the task to run concurrently with the test.
		go task.Run()

		// Wait for the job to complete or timeout.
		// This select block waits for the job to signal completion or for a timeout.
		select {
		case <-doneCh:
			select {
			// Attempt to receive from the `stopCh` channel to check if it's closed.
			// In Go, when receiving from a closed channel, the operation will return the zero value
			// of the channel's type immediately and the second value (`ok`) will be false.
			// This behavior allows us to determine if the channel has been closed by checking the `ok` value.
			case <-task.stopCh:
				// Attempt to receive from the `stopCh` channel to check if it's closed.
				// In Go, receiving from a closed channel returns the zero value immediately and `ok` is false.
				// If the channel is still open, `ok` would be true, indicating that the task is still running.
				_, ok := <-task.stopCh

				// Assert that `ok` is false, meaning that the `stopCh` should be closed at this point.
				// A closed `stopCh` indicates that the task has completed its execution and signaled completion.
				// If the channel is still open (`ok` is true), this would imply the task has not finished properly, and the test should fail.
				assert.False(t, ok, "Expected stop channel to be closed, indicating job completion")
			default:
				// The `default` case is executed if none of the other cases in the select statement are ready.
				// This provides a non-blocking path, ensuring that the select statement can proceed
				// without being stuck waiting for an input from the channels. In this context, it effectively
				// does nothing and allows the test to continue without blocking.
			}

			// Set an error on the job instance for testing purposes.
			// Here we are not setting an error, so we expect no error.
			// This ensures that the job's error handling logic is functioning as expected.
			err := task.GetError()

			// Assert that there is no error associated with the job.
			// This checks that the job's error handling logic does not report any error.
			// The assertion will fail if job.GetError() returns a non-nil error.
			assert.NoError(t, err, "Expected no error to be reported by job.GetError()")

		case <-time.After(timeout):
			// If the task takes longer than the timeout duration, the test should fail.
			// This indicates that the task did not complete in time, which is a test failure.
			t.Fatal("Timeout waiting for task to complete")
		}
	})

	// ProcessingWithPanic tests the behavior of the task processing when the processing function
	// is designed to trigger a panic. This test ensures that the task properly handles and reports
	// the panic error. The test verifies that the error handling mechanism works correctly,
	// capturing and reporting the panic with the expected error message. Additionally, it checks
	// if the task correctly signals completion or if a timeout occurs.
	t.Run("ProcessingWithPanic", func(t *testing.T) {
		// Define the worker timeout duration for the test.
		// This is the maximum amount of time we allow for the task to complete.
		timeout := 3 * time.Second

		// Initialize input data for the processing function.
		// `inputProcessingData` represents an example integer input (in this case, `222`)
		// that will be passed to the processing function. The integer is cast to `int32`
		// to match the expected data type used by the processing function.
		inputProcessingData := int32(222)

		// Define the error handler input as a string.
		// `inputErrorHandler` is set to "error handler" and is intended to be used
		// as the input for the error handling function in the task.
		// This string will simulate or represent an identifier or message to be processed by the error handler.
		inputErrorHandler := "error handler"

		// Create a mock processing task that simulates a panic during processing.
		// The mock is designed to trigger a panic to test the task's error handling.
		mockProcessingWithPanic := &MockProcessingWithPanic{}

		// Create a new task instance with the specified timeout, name, processing function, and input data.
		// This task simulates a job that will be processed within the test.
		task := NewTask(timeout, "test-task", mockProcessingWithPanic, inputProcessingData, inputErrorHandler)
		// Assert that the task was successfully created.
		// If the task is nil, it indicates a problem with task initialization.
		assert.NotNil(t, task, "Expected task to be initialized, but it was nil")

		// Create a buffered done channel to signal job completion.
		// This channel will be used to notify when the job is done.
		doneCh := make(chan struct{}, 1)
		// Set the done channel for the task using the SetDoneChannel method.
		// The method should return no error if the done channel is valid.
		_ = task.SetDoneChannel(doneCh)

		// Create a wait group to synchronize job completion.
		// The wait group will be used to wait for the task to complete.
		wg := &sync.WaitGroup{}
		// Assign the wait group to the job instance.
		// This allows the task to signal completion to the wait group.
		_ = task.SetWaitGroup(wg)

		// Create a context without a timeout to use for the task.
		// The context is used for task processing and cancellation.
		ctx := context.Background()
		// Set the parent context of the task to the newly created context.
		// This context will be used in task processing.
		_ = task.SetContext(ctx)

		// Increment the WaitGroup counter by 1.
		// This indicates that there is a new goroutine (task) that needs to be waited on.
		// The WaitGroup counter must be incremented for each goroutine that will be started,
		// so that the main test logic can properly wait for all of them to complete.
		wg.Add(1)

		// Start a new goroutine to run the task concurrently.
		// Goroutines allow tasks to execute in parallel with other operations, making it possible
		// to simulate concurrent processing and handle asynchronous operations within tests.
		go func() {
			// Introduce a delay of 1 second before executing the task.
			// This simulates a scenario where there is a delay before the task starts processing.
			// The delay ensures that the task runs after a brief wait, allowing for other operations
			// or conditions to be set up beforehand.
			<-time.After(1 * time.Second)

			// Run the task in the separate goroutine.
			// This invokes the `Run()` method of the task, which starts the task's processing logic.
			// Running the task in a separate goroutine allows it to execute concurrently with
			// other operations, such as waiting for completion or handling timeouts.
			task.Run() // Begin the task processing logic in the background.
		}()

		select {
		case <-doneCh:
			// Retrieve any error that may have been recorded by the task.
			// This checks whether the task encountered an error during its execution,
			// such as a panic or other runtime error, and allows for verification of error handling.
			err := task.GetError()

			// Assert that the mock processing function panics with the expected error message.
			// This ensures that the panic is properly triggered and that the error message matches the expected value.
			// The test checks if calling the `Processing` method on `mockProcessingWithPanic`
			// with nil context and input results in a panic with the message "mock panic".
			assert.PanicsWithError(t, "mock panic", func() {
				// Call the `Processing` method on the mock, which is expected to panic.
				// This will simulate an error scenario where the processing function fails.
				mockProcessingWithPanic.Processing(nil, nil)
			}, "Expected panic with message 'mock panic' not triggered")

			// Assert that the error retrieved from the task matches the expected error message.
			// This verifies that the error captured during task execution (if any) is exactly the one expected.
			// The error message should match "mock panic" to confirm that the task handled the panic correctly.
			assert.EqualError(t, err, "mock panic", "Task did not capture the expected panic error")
		case <-time.After(timeout):
			// If no signal is received from `stopCh` within `2 * timeout` duration, this case will trigger, indicating a timeout.
			// This is a safeguard to ensure the test doesn't hang indefinitely if something goes wrong.
			t.Fatal("timeout waiting for job to complete")
		}
	})

	// RecallingStop tests the behavior of the Stop method on the Task instance, specifically
	// focusing on ensuring that the stop channel (stopCh) is properly closed and that calling
	// Stop multiple times behaves as expected. The test verifies that the channel is closed
	// correctly, allowing subsequent reads to be non-blocking, and checks that the Stop method
	// can be safely called multiple times without causing issues.
	t.Run("RecallingStop ", func(t *testing.T) {
		// Create a new Task instance with a stop channel (stopCh) initialized.
		// The stop channel is used to signal when the task should stop processing.
		task := &Task{stopCh: make(chan struct{}, 1)}

		// Declare a sync.WaitGroup to synchronize the completion of a goroutine.
		// The WaitGroup is used to ensure that the main function waits for the goroutine to finish.
		var wg sync.WaitGroup
		// Increment the WaitGroup counter by 1.
		// This indicates that there is one goroutine that the test needs to wait for.
		wg.Add(1)

		// Start a new goroutine to simulate concurrent task behavior.
		// This goroutine will block until it receives a signal from the stop channel.
		go func() {
			// Ensure that the WaitGroup counter is decremented when the goroutine completes.
			// This allows the main test function to know when the goroutine has finished its execution.
			defer wg.Done()

			// Wait for a signal from the stop channel (stopCh).
			// The goroutine will block here until the stop channel receives a signal, indicating that the task is stopping.
			<-task.stopCh
		}()

		// Call the Stop method on the task to signal that it should stop processing.
		// This sends a signal through the stop channel to indicate that the task should stop.
		task.Stop()

		// Wait for the goroutine to finish its execution.
		// The WaitGroup will block until the goroutine calls Done(), ensuring that the goroutine has received the stop signal.
		wg.Wait()

		// Use a select statement to check the state of the stop channel after the Stop method is called.
		// This is done to verify that the stop channel has been properly closed by the Stop method.
		select {
		case <-task.stopCh:
			// Success, stopCh was closed as expected.
			// The case branch will be executed if the stop channel is closed, indicating that the Stop method worked correctly.
		default:
			// Fail the test if the stop channel is still open.
			// If the stop channel is not closed, the test will fail with this message, indicating that the Stop method did not work as expected.
			t.Fatal("stopCh should be closed, but it's not")
		}

		// Call Stop again to ensure that subsequent calls to Stop are handled correctly.
		// This tests that the Stop method correctly handles multiple invocations, especially if there are safeguards in place (e.g., using sync.Once).
		task.Stop()
	})

	// TaskStop tests the behavior of the Task's Stop method during execution.
	// This test verifies that the task correctly stops when the Stop method is called,
	// that the stop channel (`stopCh`) is properly closed, and that the task's processing
	// logic behaves as expected when interrupted by a stop signal.
	t.Run("TaskStop", func(t *testing.T) {
		// Define the worker timeout duration for the test.
		// This is the maximum amount of time we allow for the task to complete.
		timeout := 5 * time.Second

		// Initialize input data for the processing function.
		// `inputProcessingData` represents an example integer input (in this case, `222`)
		// that will be passed to the processing function. The integer is cast to `int32`
		// to match the expected data type used by the processing function.
		inputProcessingData := int32(222)

		// Define the error handler input as a string.
		// `inputErrorHandler` is set to "error handler" and is intended to be used
		// as the input for the error handling function in the task.
		// This string will simulate or represent an identifier or message to be processed by the error handler.
		inputErrorHandler := "error handler"

		// Create an instance of the mock processing task with a specified timeout.
		// This mock simulates a long-running task for testing purposes.
		mockProcessingWithLongTask := &MockProcessingLongTask{timeout: timeout}

		// Create a new task instance with the specified timeout, name, processing function, and input data.
		// This task simulates a job that will be processed within the test.
		task := NewTask(timeout, "test-task", mockProcessingWithLongTask, inputProcessingData, inputErrorHandler)
		// Assert that the task was successfully created.
		// If the task is nil, it indicates a problem with task initialization.
		assert.NotNil(t, task, "Expected task to be initialized, but it was nil")

		// Create a wait group to synchronize job completion.
		// The wait group will be used to wait for the task to complete.
		wg := &sync.WaitGroup{}
		// Assign the wait group to the job instance.
		// This allows the task to signal completion to the wait group.
		_ = task.SetWaitGroup(wg)

		// Create a context without a timeout to use for the task.
		// The context is used for task processing and cancellation.
		ctx := context.Background()
		// Set the parent context of the task to the newly created context.
		// This context will be used in task processing.
		_ = task.SetContext(ctx)

		// Increment the WaitGroup counter by 1.
		// This indicates that there is a new goroutine (task) that needs to be waited on.
		// The WaitGroup counter must be incremented for each goroutine that will be started,
		// so that the main test logic can properly wait for all of them to complete.
		wg.Add(1)

		// Start a new goroutine to run the task concurrently.
		// Goroutines allow tasks to execute in parallel with other operations, making it possible
		// to simulate concurrent processing and handle asynchronous operations within tests.
		go func() {
			// Introduce a delay of 1 second before executing the task.
			// This simulates a scenario where there is a delay before the task starts processing.
			// The delay ensures that the task runs after a brief wait, allowing for other operations
			// or conditions to be set up beforehand.
			<-time.After(1 * time.Second)

			// Run the task in the separate goroutine.
			// This invokes the `Run()` method of the task, which starts the task's processing logic.
			// Running the task in a separate goroutine allows it to execute concurrently with
			// other operations, such as waiting for completion or handling timeouts.
			task.Run() // Begin the task processing logic in the background.
		}()

		// Wait for a brief moment (1 second) before stopping the task.
		// This simulates a scenario where the task is stopped shortly after it starts.
		// It introduces a small delay to allow the task to start processing before being asked to stop.
		<-time.After(1 * time.Second)

		// Gracefully stop the task by calling the `Stop()` method.
		// This sends a signal through the `stopCh` channel to indicate that the task should halt its execution.
		task.Stop()

		// Use a select statement to either receive a signal from the stop channel or timeout.
		// This ensures that we properly handle the task's stopping behavior and confirm the channel's closure.
		select {
		case _, ok := <-task.stopCh:
			// Attempt to receive a value from the `stopCh` channel.
			// The `ok` variable will be false if the channel is closed, which is the expected behavior after `Stop()` is called.
			assert.False(t, ok, "Expected stopCh to be closed, but it was not")

			// Introduce a short delay to ensure that asynchronous operations have time to complete.
			// This delay allows the stop signal to propagate and any remaining processing to finalize.
			<-time.After(10 * time.Millisecond)

			// Assert that the counter in the mock processing task matches the expected value.
			// This verifies that the task was properly stopped and that the mock processing function was executed as expected.
			assert.Equal(t, MockProcessingLongTaskCounter, mockProcessingWithLongTask.Counter(), "Task processing counter did not match expected value after Stop() was called")
		case <-time.After(2 * timeout):
			// If no signal is received from `stopCh` within `2 * timeout` duration, this case will trigger, indicating a timeout.
			// This is a safeguard to ensure the test doesn't hang indefinitely if something goes wrong.
			t.Fatal("timeout waiting for job to complete")
		}

	})
}

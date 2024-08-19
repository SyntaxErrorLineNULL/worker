package worker

import (
	"context"
	"sync"
	"testing"
	"worker/mocks"

	"github.com/stretchr/testify/assert"
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
}

package worker

import (
	"context"
	"testing"
	wr "worker"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Parallel()

	// InitWorker tests the initialization of a Worker instance.
	// It verifies that the worker is properly created with the expected initial values
	// and that all required channels are initialized. This test ensures that
	// the worker is correctly set up with the specified ID, timeout, and initial status.
	t.Run("InitWorker", func(t *testing.T) {
		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(1)
		// Assert that the worker instance is not nil.
		// This checks that the worker was successfully created and is not a zero value.
		assert.NotNil(t, worker, "Worker should be successfully created")

		// Assert that the worker's ID is set to 1.
		// This verifies that the worker has been initialized with the correct ID.
		assert.Equal(t, int64(1), worker.workerID, "Worker ID should be 1")

		// Assert that the worker's status is initially set to `StatusWorkerIdle`.
		// This confirms that the worker starts in an idle state, ready to process jobs.
		assert.Equal(t, wr.StatusWorkerIdle, worker.GetStatus(), "Worker status should be idle initially")

		// Assert that the worker's stop channel is not nil.
		// This checks that the stop channel is correctly created as part of the worker's initialization.
		assert.NotNil(t, worker.stopCh, "Worker stop channel should be created")

		// Assert that the worker's error channel is not nil.
		// This verifies that the error channel is correctly initialized, allowing the worker to report errors.
		assert.NotNil(t, worker.errCh, "Worker error channel should be created")
	})

	// SetQueue tests the SetQueue method of the Worker type.
	// It verifies that the method correctly handles setting both open and closed channels as the task queue.
	// The test ensures that setting an open channel succeeds, while setting a closed channel produces an error,
	// validating proper queue management within the worker.
	t.Run("SetQueue", func(t *testing.T) {
		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(1)
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
		// Create a new Worker instance with ID 1, a timeout of 1 second, and a logger.
		// This initializes the worker with specified parameters and ensures that it is properly set up.
		worker := NewWorker(1)
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

		// Attempt to set the worker's context to nil.
		// This tests the SetContext method's handling of invalid input.
		err = worker.SetContext(nil)
		// Assert that an error occurred when trying to set a nil context.
		// This verifies that the method correctly returns an error for invalid input.
		assert.Error(t, err, "Expected an error when setting a nil context")

		// Assert that the worker's context remains unchanged after attempting to set a nil context.
		// This verifies that the context does not get altered if a nil value is provided.
		assert.Equal(t, ctx, worker.workerContext, "Worker context should remain unchanged when setting a nil context")
	})
}

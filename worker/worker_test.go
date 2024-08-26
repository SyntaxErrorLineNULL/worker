package worker

import (
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
}

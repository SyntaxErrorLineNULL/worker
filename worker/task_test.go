package worker

import (
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
		// Create a new instance of Job.
		job := &Task{}

		// Set the wait group for the job using the SetWaitGroup method.
		err := job.SetWaitGroup(&wg)
		// Assert that no error was returned from SetWaitGroup.
		// This check confirms that the method handled the valid WaitGroup correctly.
		assert.NoError(t, err, "expected no error when setting a valid WaitGroup")

		// Assert that the wait group was set correctly in the job.
		// The expected value is the address of the initialized WaitGroup.
		// The actual value is the job's wg field.
		// The message "expected the job's wait group to be set correctly" is displayed if the assertion fails.
		assert.Equal(t, &wg, job.wg, "expected the job's wait group to be set correctly")
	})

	// SetDoneChannel tests the SetDoneChannel method of the Task struct.
	// It ensures that the method correctly assigns a provided done channel to the Task,
	// and verifies that no errors occur during this process.
	t.Run("SetDoneChannel", func(t *testing.T) {
		// Create a done channel.
		doneCh := make(chan struct{})

		// Create a new instance of Job.
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
}

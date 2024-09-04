package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/SyntaxErrorLineNULL/worker"
)

// Task is a struct that encapsulates the execution logic of a concurrent task.
// It manages task execution, including timeout, error handling, and coordination
// with external components through contexts, channels, and WaitGroups.
type Task struct {
	// parentCtx is the parent context passed to the task.
	// This context is used as the base context for the task's execution and can be used to propagate cancellation signals.
	parentCtx context.Context
	// name is the name of the task, primarily used for logging purposes.
	// It helps in identifying and tracking the task in logs.
	name string
	// timeout is the maximum execution time allowed for "long" tasks.
	// If the task takes longer than this duration, it will be forcefully stopped.
	timeout time.Duration
	// processing is a module for processing the task we received, each module implements the Processing interface,
	// Task receives the task and the necessary module to run it.
	processing worker.Processing
	// processingInput holds the input data required by the processing module to execute the task.
	processingInput interface{}
	// doneCh is a signaling channel used to notify external handlers when the task is complete.
	// It's an optional channel, primarily used for coordinating task completion with other processes.
	doneCh chan<- struct{}
	// wg is a WaitGroup that is used to wait for the task to complete.
	// It is only necessary for tasks that require controlled execution, ensuring that all parts of the task finish before proceeding.
	wg *sync.WaitGroup
	// stopCh this is channel for signaling task termination.
	// In case of uncontrolled task's timeout <= 0 will not work, in case of controlled tasks it
	// will close the context and the handler will not be able to do its task.
	stopCh chan struct{}
	// stopOnce is used to ensure that the stopCh is closed only once.
	// This is important to prevent multiple close operations, which could cause a panic.
	stopOnce sync.Once
	// err is a channel to receive error messages, in case of panic triggers
	// it will help prevent processing crash by simply passing the error to the worker for logging.
	err error
}

// NewTask initializes a new Task instance with the provided parameters.
// The function takes the maximum timeout, task name, processing module, and inputs for processing and error handling.
func NewTask(timeout time.Duration, taskName string, processing worker.Processing, processingInput interface{}) *Task {
	return &Task{
		timeout:         timeout,
		name:            taskName,
		processing:      processing,
		processingInput: processingInput,
		err:             nil,
		stopCh:          make(chan struct{}, 1),
	}
}

// SetWaitGroup assigns the provided WaitGroup to the task's wg field.
// This allows the task to signal when it is done by using the WaitGroup.
func (t *Task) SetWaitGroup(wg *sync.WaitGroup) error {
	// Check if the provided WaitGroup is nil.
	// A nil WaitGroup would cause a panic if used to signal task completion.
	// Return an error to prevent assigning an invalid WaitGroup.
	if wg == nil {
		return errors.New("WaitGroup cannot be nil")
	}

	// Assign the provided non-nil WaitGroup to the task's wg field.
	t.wg = wg

	// Return nil to indicate that the WaitGroup was successfully set.
	return nil
}

// SetDoneChannel sets the provided channel as the done channel for the task.
// It first checks if the provided channel is nil or already closed to prevent
// setting an invalid channel, which could lead to runtime errors.
//
// Note: the channel must be closed after you receive a signal that the process is complete.
func (t *Task) SetDoneChannel(done chan struct{}) error {
	// Check if the provided channel is nil. A nil channel is invalid and
	// cannot be used, so return an error in this case.
	if done == nil {
		return worker.ChanIsEmptyError
	}

	// Use a non-blocking select statement to check if the channel is closed.
	// This ensures that the method can detect a closed channel without blocking
	// the execution, allowing for safe channel assignment.
	select {
	case <-done:
		// If this case is executed, it means the channel has already been closed.
		// Return an error indicating that a closed channel cannot be set.
		return worker.ChanIsCloseError
	default:
		// If the channel is not closed (i.e., it's still open), proceed to set it.
		// This case will execute immediately if the channel is open.
	}

	// Assign the provided open channel to the task's doneCh field. This allows
	// the task to later use this channel for signaling that it is done.
	t.doneCh = done

	// Return nil to indicate that the channel was successfully set.
	return nil
}

// SetContext sets the parent context for the task.
// This allows updating the parent context during the execution of the task.
func (t *Task) SetContext(ctx context.Context) error {
	// Check if the provided context is nil. A nil context is invalid and
	// should not be used. Return an error in this case.
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	// Assign the provided valid context to the task's parentCtx field.
	t.parentCtx = ctx

	// Return nil to indicate that the context was successfully set.
	return nil
}

// GetError returns a channel that can be used to receive error messages.
// In case of panic, the worker will be able to log the error in his or her own account,
// which can help when looking for problems.
func (t *Task) GetError() error {
	return t.err
}

// GetName returns the name of the task.
// This method retrieves the name assigned to the task instance.
func (t *Task) String() string {
	return t.name
}

// Run orchestrates the execution of a task with proper lifecycle management using contexts.
// It handles both short and long-running tasks, manages execution time through context timeouts,
// and ensures proper cleanup and error handling.
func (t *Task) Run() {
	// Defer cleanup steps.
	defer func() {
		// Recover from any panic that occurs during the task's execution.
		// This ensures that the task does not crash the entire application if a panic occurs.
		if rec := recover(); rec != nil {
			// Convert the recovered panic into an error using a helper function.
			// This step captures the panic and translates it into an error that can be handled.
			if err := worker.GetRecoverError(rec); err != nil {
				// Store the error in the task for later retrieval.
				t.err = err
			}

		}

		// Call the Stop method to ensure that the task is properly stopped.
		// This includes closing the stop channel stopCh and performing any necessary cleanup.
		t.Stop()

		// Signal the completion of the task to the doneCh` channel, if it is available.
		// Sending an empty struct on doneCh notifies any listeners that the task has finished.
		if t.doneCh != nil {
			t.doneCh <- struct{}{}
		}

		// Decrement the wait group wg counter, if it is available.
		// This signals that the task's goroutine has completed its work, allowing any waiting processes to proceed.
		if t.wg != nil {
			t.wg.Done()
		}
	}()

	// Check if the task has no timeout set (i.e., timeout is zero or negative).
	if t.timeout <= 0 {
		// Start the task's processing logic in a new Goroutine to allow it to run concurrently with other tasks.
		go t.processing.Processing(t.parentCtx, t.processingInput)
		// Since there's no timeout, return from the Run method immediately after launching the goroutine.
		return
	}

	// Create a new context with a timeout based on workerTimeout.
	taskContext, cancel := context.WithTimeout(t.parentCtx, t.timeout)

	// Initialize a done channel with a buffer of 1, which will be used to signal when the task's processing is complete.
	done := make(chan struct{}, 1)

	// If a WaitGroup (wg) is associated with this task, increment its counter to track the new goroutine that will be started.
	if t.wg != nil {
		// Increase the counter in the wait group to indicate that another goroutine is starting.
		t.wg.Add(1)
	}

	// Start a Goroutine to execute the main function with the created context and done channel.
	go func() {
		defer func() {
			// Decrement the wait group to signal task completion.
			if t.wg != nil {
				t.wg.Done()
			}
		}()

		// Call the initProcessing method to start the task's main processing logic, passing the context and done channel.
		// If an error occurs during processing, store it in the task's error field.
		if err := t.initProcessing(taskContext, done); err != nil {
			// Store the error that occurred during task processing.
			t.err = err
		}
	}()

	// Select block to wait for the completion of the main function or the timeout.
	select {
	// A signal was received from the stop channel stopCh, indicating that the task should stop its execution.
	// This signal is typically sent when the Stop method is called, which means the task is being explicitly requested to terminate.
	case <-t.stopCh:
		fmt.Println("\nStop ch")
		// Cancel the context to immediately stop any ongoing or pending operations associated with this task.
		// This ensures that the task's processing halts as soon as possible.
		cancel()

		// Return from the `Run` method, terminating the task's execution.
		// Exiting the method signals that the task has acknowledged the stop request and is no longer running.
		return

	// A signal was received from the done channel, indicating that the primary function has completed its execution.
	// This means the task finished its processing before the timeout or stop signal occurred.
	case <-done:
		// Close the done channel to signal that no more events will occur on this channel.
		// Closing the channel is an important step to clean up resources and prevent any further sends on this channel.
		close(done)
		// Cancel the context to stop any remaining work associated with this task.
		// This is a cleanup step to ensure that no further actions are taken by the task now that it has completed.
		cancel()
		// Return from the Run method, as the task has successfully completed.
		// Exiting the method indicates that the task's lifecycle is finished.
		return

	// If the task's context times out or is canceled, this case will be triggered.
	// This happens when the task either exceeds the allowed time (timeout) or if the parent context is canceled.
	case <-taskContext.Done():
		// If the task's context times out or is canceled, this case will be triggered.
		// Cancel the task's context to halt any ongoing processing.
		cancel()
	}
}

// initProcessing initializes the processing of the job within the given context.
// This method is needed to prevent overlapping of panic catching. It starts processing
// in this method, and in case of an error, the local recover will be triggered.
// The global error processing will be triggered in the Run method.
func (t *Task) initProcessing(ctx context.Context, doneCh chan struct{}) (err error) {
	defer func() {
		// Recover from any panic that occurs during the task's execution.
		// This ensures that the task does not crash the entire application if a panic occurs.
		if rec := recover(); rec != nil {
			// Convert the recovered panic into an error using a helper function.
			// This step captures the panic and translates it into an error that can be handled.
			err = worker.GetRecoverError(rec)
		}

		// Notify that the task has completed, whether successfully or after a panic.
		// Sending an empty struct to `doneCh` signals that the task's execution is done.
		doneCh <- struct{}{}
	}()

	// start a task with a context that will be terminated after timeout,
	// and processing will not be able to perform any further actions.
	// Then ErrorHandler with a new context will be triggered, it acts as a compensating action.
	t.processing.Processing(ctx, t.processingInput)

	// Default return value, indicating error stat.
	return
}

// Stop gracefully stops the job by signaling through the stop channel and closing it.
// This method ensures that the job's stop channel is only closed once, even if Stop is called multiple times.
// It uses sync.Once to guarantee that the channel is closed exactly once, preventing multiple closure attempts
// which could lead to a panic or undefined behavior.
func (t *Task) Stop() {
	// Execute the stop logic only once, even if Stop is called multiple times.
	t.stopOnce.Do(func() {
		if t.stopCh != nil {
			// Send an empty struct to the stop channel to signal that the job should stop.
			// This notification allows any goroutines waiting on this channel to detect that the job is stopping.
			t.stopCh <- struct{}{}

			// Close the stop channel to indicate that no more signals will be sent.
			// Closing the channel is important for cleanup and to signal all listeners that no further events will occur.
			close(t.stopCh)
		}
	})
}

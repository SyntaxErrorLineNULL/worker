package worker

import (
	"context"
	"errors"
	"sync"
	"time"
	"worker"
)

type Task struct {
	// parentCtx is the parent context passed to the job.
	parentCtx context.Context
	// Job name, for logging
	name string
	// timeout is the maximum execution time for "long" tasks.
	timeout time.Duration
	// processing is a module for processing the task we received, each module implements the Processing interface,
	// Task receives the task and the necessary module to run it.
	processing worker.Processing

	processingInput             interface{}
	processingErrorHandlerInput interface{}

	processingOutputCh chan interface{}
	processingErrorCh  chan error

	processingErrorHandlerOutputCh chan interface{}
	processingErrorHandlerErrorCh  chan error

	// doneCh this is signaling channel for external handlers about job completion
	doneCh chan<- struct{}
	// wg is a WaitGroup to wait for the job to complete.
	// only necessary for controlled tasks
	wg *sync.WaitGroup
	// stopCh this is channel for signaling job termination.
	// In case of uncontrolled task's timeout <= 0 will not work, in case of controlled tasks it
	// will close the context and the handler will not be able to do its job.
	stopCh chan struct{}
	// stopOnce ensures the stopCh is closed only once.
	stopOnce sync.Once
	// err is a channel to receive error messages, in case of panic triggers
	// it will help prevent processing crash by simply passing the error to the worker for logging.
	err error
}

func NewTask(timeout time.Duration, taskName string, processing worker.Processing, processingInput, errorHandlerInput interface{}) *Task {
	return &Task{
		timeout:                        timeout,
		name:                           taskName,
		processing:                     processing,
		processingInput:                processingInput,
		processingErrorHandlerInput:    errorHandlerInput,
		processingOutputCh:             make(chan interface{}, 1),
		processingErrorCh:              make(chan error, 1),
		processingErrorHandlerOutputCh: make(chan interface{}, 1),
		processingErrorHandlerErrorCh:  make(chan error, 1),
		// def task error
		err:    nil,
		stopCh: make(chan struct{}, 1),
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
func (t *Task) SetDoneChannel(done chan struct{}) error {
	// Check if the provided channel is nil. A nil channel is invalid and
	// cannot be used, so return an error in this case.
	if done == nil {
		return errors.New("channel cannot be nil")
	}

	// Use a non-blocking select statement to check if the channel is closed.
	// This ensures that the method can detect a closed channel without blocking
	// the execution, allowing for safe channel assignment.
	select {
	case <-done:
		// If this case is executed, it means the channel has already been closed.
		// Return an error indicating that a closed channel cannot be set.
		return errors.New("cannot set a closed channel")
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

// GetName returns the name of the task.
// This method retrieves the name assigned to the task instance.
func (t *Task) String() string {
	return t.name
}

func (t *Task) Run() {
	// Defer cleanup steps.
	defer func() {
		// Recover from any panic in the job and report it.
		if rec := recover(); rec != nil {
			err := worker.GetRecoverError(rec)
			if err != nil {
				t.err = err
			}
		}

		return
	}()

	// Defer function to signal job completion and decrement the wait group.
	defer func() {
		// Signal the completion to the doneCh channel if available.
		if t.doneCh != nil {
			t.doneCh <- struct{}{}
		}

		// Decrement the wait group if available.
		if t.wg != nil {
			t.wg.Done()
		}

		return
	}()

	// Synchronize execution of goroutines using a done channel.
	done := make(chan struct{}, 1)

	if t.timeout <= 0 {
		// Add 1 to the wait group to track the execution of the main function.
		t.wg.Add(1)

		go func() {
			// Initialize the processing with the provided context.
			if err := t.initProcessing(t.parentCtx, done); err != nil {
				t.err = err
			}

			return
		}()

		return
	}

	// Create a new context with a timeout based on workerTimeout.
	jobCtx, cancel := context.WithTimeout(t.parentCtx, t.timeout)
	defer cancel()

	// Add 1 to the wait group to track the execution of the main function.
	t.wg.Add(1)

	// Start a Goroutine to execute the main function with the created context and done channel.
	go func() {
		// Initialize the processing with the provided context.
		if err := t.initProcessing(jobCtx, done); err != nil {
			t.err = err
		}
		return
	}()

	// Select block to wait for the completion of the main function or the timeout.
	select {
	case <-done:
		// Primary function completed within the allowed time.
		// Close the done channel and continue with the rest of your program.
		close(done)
		return

	case <-jobCtx.Done():
		// Why a new context and not use j.parentCtx / jobCtx?
		// The point is that if the handler context is completed, but Processing itself is not completed,
		// ErrorHandler is called, and it can execute its logic with the new context.
		errorHandlerResult, errorHandlerErr := t.processing.ErrorHandler(context.Background(), t.processingErrorHandlerInput)

		t.processingErrorHandlerOutputCh <- errorHandlerResult
		t.processingErrorHandlerErrorCh <- errorHandlerErr

		return
	}
}

// initProcessing initializes the processing of the job within the given context.
// This method is needed to prevent overlapping of panic catching. It starts processing
// in this method, and in case of an error, the local recover will be triggered.
// The global error processing will be triggered in the Run method.
func (t *Task) initProcessing(ctx context.Context, done chan struct{}) (err error) {
	defer func() {
		// Recover from any panic in the job and report it.
		if rec := recover(); rec != nil {
			// Convert the recovered panic into an error using internal.GetRecoverError.
			err = worker.GetRecoverError(rec)
			return
		}

		// Return from the function. Default value of err is nil if no panic occurs.
		return
	}()

	defer func() {
		// Decrement the wait group to signal task completion.
		t.wg.Done()
		// the task has been completed successfully, which means that we will notify the handler about it.
		done <- struct{}{}
		return
	}()

	// start a task with a context that will be terminated after timeout,
	// and processing will not be able to perform any further actions.
	// Then ErrorHandler with a new context will be triggered, it acts as a compensating action.
	processingResult, processingError := t.processing.Processing(ctx, t.processingInput)

	t.processingOutputCh <- processingResult
	t.processingErrorCh <- processingError

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
		// Send an empty struct to the stop channel to signal that the job should stop.
		// This notification allows any goroutines waiting on this channel to detect that the job is stopping.
		t.stopCh <- struct{}{}

		// Close the stop channel to indicate that no more signals will be sent.
		// Closing the channel is important for cleanup and to signal all listeners that no further events will occur.
		close(t.stopCh)

		return
	})
}

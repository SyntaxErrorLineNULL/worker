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

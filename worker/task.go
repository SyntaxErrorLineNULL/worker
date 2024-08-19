package worker

import (
	"context"
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

// GetName returns the name of the task.
// This method retrieves the name assigned to the task instance.
func (t *Task) String() string {
	return t.name
}

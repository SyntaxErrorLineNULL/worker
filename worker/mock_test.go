//nolint:all
package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	wr "github.com/SyntaxErrorLineNULL/worker"
)

var (
	// errMockPanic is a predefined error used to simulate a panic scenario in tests.
	// This can be used to test how the system handles unexpected panics during task processing.
	errMockPanic = errors.New("mock panic")
)

// MockProcessingWithPanic is a mock implementation of the Processing interface.
// It is used in tests to simulate scenarios where the Processing method
// deliberately causes a panic to test error handling and recovery mechanisms.
type MockProcessingWithPanic struct{}

// Processing simulates a processing operation and deliberately causes a panic.
// This method is used to test the behavior of the system when a panic occurs
// during processing. It returns false, but the primary purpose is to trigger
// a panic with a predefined error message to test panic recovery mechanisms.
func (m *MockProcessingWithPanic) Processing(_ context.Context, _ interface{}) {
	// This mock implementation deliberately causes a panic with a specific error message.
	// It helps simulate and test how the system handles unexpected errors during processing.
	panic(errors.New("mock panic"))
}

// MockProcessingLongTaskResult represents the result of processing a long-running task in a mock environment.
// This structure is used in tests to verify the state of the task's context after it has been processed.
// It contains flags indicating whether the context was canceled (done) or not during task execution.
type MockProcessingLongTaskResult struct {
	// ContextIsDone indicates whether the context was properly canceled during the task processing.
	// This is set to true if the context's Done channel was closed, signaling the task to stop.
	ContextIsDone bool

	// ContextIsNotDone indicates that the context was not canceled during the task processing.
	// This is set to true if the context's Done channel remained open, meaning the task continued without interruption.
	ContextIsNotDone bool
}

// MockProcessingLongTask is a mock implementation of the Processing interface designed to simulate long-running tasks.
// It is primarily used for testing purposes to validate how a task behaves when it takes a considerable amount of time to complete.
type MockProcessingLongTask struct {
	// timeout specifies the duration for which the task should simulate processing.
	// This duration represents the "long" time that the task will take before completion.
	timeout time.Duration
}

// Processing simulates the execution of a long-running task.
// It takes a context and an input parameter (both of which are ignored in this mock implementation) and sleeps for the specified timeout duration.
// This method is used to mimic the behavior of a task that consumes time and to test how the task handling mechanism responds to such delays.
func (m *MockProcessingLongTask) Processing(ctx context.Context, input interface{}) {
	resultCh := input.(chan *MockProcessingLongTaskResult)
	select {
	case <-ctx.Done():
		fmt.Println("\nProcessing context done")
		resultCh <- &MockProcessingLongTaskResult{ContextIsDone: true}
	case <-time.After(m.timeout):
		// Simulate long processing by blocking for the duration specified in m.timeout.
		// This is done using time.After to block the goroutine until the timeout has elapsed.
		fmt.Println("\nProcessing timeout")
		resultCh <- &MockProcessingLongTaskResult{ContextIsNotDone: true}
	}
}

// MockSuccessProcessing is a mock implementation of the Processing interface
// used for testing purposes. It simulates a processing task and tracks
// the number of tasks processed.
type MockSuccessProcessing struct {
	// counter keeps track of the volume IDs processed by this mock.
	// This is used to verify that tasks are processed as expected.
	counter atomic.Int32
}

// Processing simulates the processing of a task. It increments the counter
// based on the volume ID of the task and returns true to indicate successful
// processing.
func (m *MockSuccessProcessing) Processing(_ context.Context, input interface{}) {
	// Log the initiation of the processing method for debugging purposes.
	fmt.Println("Init MockSuccessProcessing Processing")

	// Add the volume ID of the task to the counter.
	// This simulates the effect of processing a task by incrementing the counter
	// based on the volume ID provided by the task.
	m.counter.Add(input.(int32))
}

// MockPanicTask is a mock implementation of the Job interface designed to simulate
// a task that panics during execution. This can be used in testing scenarios where
// you need to verify the behavior of a worker or system when a task causes a panic.
type MockPanicTask struct{}

// SetWaitGroup simulates setting a wait group for the job.
// It immediately marks the wait group as done, as if the task has completed.
func (t *MockPanicTask) SetWaitGroup(wg *sync.WaitGroup) error {
	// Mark the wait group as done.
	wg.Done()
	return nil
}

// SetDoneChannel is a mock implementation that does nothing with the done channel.
// It simply returns nil, indicating success in setting the channel.
func (t *MockPanicTask) SetDoneChannel(_ chan struct{}) error {
	// No operation needed, simply return nil.
	return nil
}

// SetContext is a mock implementation that does nothing with the provided context.
// It simply returns nil, indicating success in setting the context.
func (t *MockPanicTask) SetContext(_ context.Context) error {
	// No operation needed, simply return nil.
	return nil
}

// GetError is a mock implementation that returns nil.
// This simulates a task that does not encounter any error.
func (t *MockPanicTask) GetError() <-chan error {
	// No error, so return nil.
	return nil
}

// String returns an empty string as a mock representation of the task.
func (t *MockPanicTask) String() string {
	// Return an empty string to represent the job.
	return ""
}

// Run simulates the execution of the task and intentionally causes a panic.
// This is used to test how the system handles a job that panics.
func (t *MockPanicTask) Run(_ time.Duration) {
	// Simulate a panic occurring during job execution.
	panic(errMockPanic)
}

// Stop is a mock implementation that does nothing.
// It simulates stopping the task, though no operation is performed here.
func (t *MockPanicTask) Stop() {
	// No operation needed for stopping.
}

// MockWorkerWithPanic simulates a worker that panics during its execution.
// This mock is used to test how the system handles workers that encounter
// panic situations and need to be restarted. It tracks retries, restarts, and errors.
type MockWorkerWithPanic struct {
	// Context for controlling the lifecycle of the worker.
	ctx context.Context
	// Atomic counter for tracking retry attempts.
	retry atomic.Int32
	// Atomic counter for tracking restart attempts after a panic.
	restart atomic.Int32
	// Channel for sending errors that occur during the worker's operation.
	errCh chan *wr.Error
	// Channel for signaling when the worker should stop.
	stopCh chan struct{}
	// Flag indicating whether the worker has encountered a panic.
	withPanic bool
	// Channel for signaling when the worker has completed its operations.
	doneCh chan struct{}
	isStop atomic.Bool
}

// String returns a string representation of the mock worker.
// This is primarily used for logging and debugging purposes, identifying
// the worker by its type.
func (m *MockWorkerWithPanic) String() string {
	return "MockWorkerWithPanic"
}

// SetContext sets the context for the worker.
// This allows external control over the worker's lifecycle through the provided context.
// The context can be used to cancel or time out the worker's execution.
func (m *MockWorkerWithPanic) SetContext(ctx context.Context) error {
	m.ctx = ctx
	return nil
}

// SetQueue sets the task queue for the worker.
// In this mock implementation, the method does nothing but exists to fulfill
// the worker interface contract. In a real worker, this would allow the worker
// to pull jobs from the queue.
func (m *MockWorkerWithPanic) SetQueue(queue chan wr.Task) error {
	return nil
}

// SetWorkerErrChannel sets the error channel for the worker.
// This channel is used to communicate panic or other serious errors
// back to the worker pool or controlling system. The worker sends an error
// when it encounters a panic situation.
func (m *MockWorkerWithPanic) SetWorkerErrChannel(errCh chan *wr.Error) error {
	m.errCh = errCh
	return nil
}

// Restart increments the retry counter and restarts the worker.
// This method simulates restarting the worker after it has encountered an error or panic.
// The retry counter is incremented to track the number of restart attempts.
func (m *MockWorkerWithPanic) Restart(wg *sync.WaitGroup) {
	// Increment the retry counter.
	m.retry.Add(1)
	// Start the worker again.
	m.Start(wg)
}

// Start initiates the worker's operation.
// If the worker has not yet panicked, it waits for 2 seconds before simulating a panic.
// After the panic, the worker sends an error message on the error channel and stops.
// If the worker has already panicked, it stops and signals its completion.
func (m *MockWorkerWithPanic) Start(wg *sync.WaitGroup) {
	if !m.withPanic {
		// Simulate normal operation before the panic occurs.
		time.Sleep(2 * time.Second)
		wg.Done()
		// Mark the worker as having encountered a panic.
		m.withPanic = true
		// Send an error to the error channel to indicate the panic.
		m.errCh <- &wr.Error{Error: errMockPanic, Instance: m}
		return
	} else {
		// Signal that the worker is stopping after the restart.
		defer wg.Done()
		// Notify that the worker has stopped.
		m.stopCh <- struct{}{}
		// Increment the restart counter to track how many times the worker has restarted.
		m.restart.Add(1)
		return
	}
}

// Stop stops the worker and returns a channel that signals completion.
// The done channel is used to notify that the worker has finished its operations.
func (m *MockWorkerWithPanic) Stop() <-chan struct{} {
	// Send a message to indicate that the worker is done.
	m.doneCh <- struct{}{}
	return m.doneCh
}

func (m *MockWorkerWithPanic) IsStop() bool {
	return m.isStop.Load()
}

// GetError returns the error channel of the worker.
// This channel is used to send panic errors when the worker encounters an issue.
func (m *MockWorkerWithPanic) GetError() chan *wr.Error {
	return m.errCh
}

// GetRetry returns the number of retry attempts made by the worker.
// This method provides access to the retry counter for testing purposes.
func (m *MockWorkerWithPanic) GetRetry() int32 {
	return m.retry.Load()
}

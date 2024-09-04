package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// errMockPanic is a predefined error used to simulate a panic scenario in tests.
	// This can be used to test how the system handles unexpected panics during task processing.
	errMockPanic = errors.New("some panic")
)

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

// MockPanicTask is a mock implementation of the Task interface designed to simulate
// a task that panics during execution. This can be used in testing scenarios where
// you need to verify the behavior of a worker or system when a task causes a panic.
type MockPanicTask struct{}

// SetWaitGroup simulates setting a wait group for the task.
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
func (t *MockPanicTask) GetError() error {
	// No error, so return nil.
	return nil
}

// String returns an empty string as a mock representation of the task.
func (t *MockPanicTask) String() string {
	// Return an empty string to represent the task.
	return ""
}

// Run simulates the execution of the task and intentionally causes a panic.
// This is used to test how the system handles a task that panics.
func (t *MockPanicTask) Run() {
	// Simulate a panic occurring during task execution.
	panic(errMockPanic)
}

// Stop is a mock implementation that does nothing.
// It simulates stopping the task, though no operation is performed here.
func (t *MockPanicTask) Stop() {
	// No operation needed for stopping.
}

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// MockProcessingLongTaskCounter is a constant used to represent the value
	// added to the contextDone counter when the processing is interrupted
	// by a context cancellation. This value is used for testing purposes to
	// verify that the task handles context cancellation correctly.
	MockProcessingLongTaskCounter = int32(100)

	// MockProcessingLongTaskErrorHandlerCounter represents a fixed value that is added
	// to the errorHandlerContextDone counter. This is used in testing to check if the error
	// handler was triggered and how it handles errors.
	MockProcessingLongTaskErrorHandlerCounter = int32(1000)
)

// MockProcessingLongTask is a mock implementation of the Processing interface designed to simulate long-running tasks.
// It is primarily used for testing purposes to validate how a task behaves when it takes a considerable amount of time to complete.
type MockProcessingLongTask struct {
	// Counter for tracking context cancellations.
	contextDone atomic.Int32
	// Counter for tracking error handler invocations.
	errorHandlerContextDone atomic.Int32
	// timeout specifies the duration for which the task should simulate processing.
	// This duration represents the "long" time that the task will take before completion.
	timeout time.Duration
}

// Processing simulates the execution of a long-running task.
// It takes a context and an input parameter (both of which are ignored in this mock implementation) and sleeps for the specified timeout duration.
// This method is used to mimic the behavior of a task that consumes time and to test how the task handling mechanism responds to such delays.
func (m *MockProcessingLongTask) Processing(ctx context.Context, _ interface{}) {
	select {
	case <-ctx.Done():
		fmt.Println("\nProcessing context done")
		// The context was canceled before the timeout elapses.
		// Increment the contextDone counter by MockProcessingLongTaskCounter to indicate interruption.
		m.contextDone.Add(MockProcessingLongTaskCounter)
		return
	case <-time.After(m.timeout):
		// Simulate long processing by blocking for the duration specified in m.timeout.
		// This is done using time.After to block the goroutine until the timeout has elapsed.
		fmt.Println("\nProcessing timeout")
	}
}

// ErrorHandler is a mock implementation of the error handling method.
// This method is a no-op (no operation) in this mock, as it is not used in the current test scenarios.
// In a real-world implementation, this method would handle any errors encountered during the processing.
func (m *MockProcessingLongTask) ErrorHandler(_ context.Context, _ interface{}) {
	// No operation in this mock implementation.
	m.errorHandlerContextDone.Add(MockProcessingLongTaskErrorHandlerCounter)
}

// Result is a mock implementation of a method that would return the result of the processing.
// In this mock, it simply returns nil because the result handling is not the focus of the test scenarios involving this mock task.
func (m *MockProcessingLongTask) Result() chan interface{} {
	// Return nil as no result handling is implemented in this mock.
	return nil
}

// Counter returns the current value of the contextDone counter.
// This method allows external code to retrieve the value of contextDone, which indicates
// how many times the task was interrupted by a context cancellation during testing.
func (m *MockProcessingLongTask) Counter() int32 {
	return m.contextDone.Load()
}

// ErrorHandlerCounter returns the current value of the `erroHandlerContextDone` counter.
// This method provides the count of times the `ErrorHandler` was invoked during the execution of
// the mock task, useful for verifying if error handling was triggered in the test scenarios.
func (m *MockProcessingLongTask) ErrorHandlerCounter() int32 {
	return m.errorHandlerContextDone.Load()
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

// ErrorHandler provides a mock implementation of an error handling function.
// In this mock, the ErrorHandler method does not perform any actual error handling.
// It is included to satisfy the interface but does not implement any functionality.
func (m *MockProcessingWithPanic) ErrorHandler(_ context.Context, _ interface{}) {
	// No actual error handling is performed in this mock implementation.
	// This method is provided to fulfill the interface requirements and
	// does not affect the test scenarios directly.
}

// Result is a mock implementation of a method that would return the result of the processing.
// In this mock, it simply returns nil because the result handling is not the focus of the test scenarios involving this mock task.
func (m *MockProcessingWithPanic) Result() chan interface{} {
	// Return nil as no result handling is implemented in this mock.
	return nil
}

var ErrorMockPanic = errors.New("some panic")

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
	panic(ErrorMockPanic)
}

// Stop is a mock implementation that does nothing.
// It simulates stopping the task, though no operation is performed here.
func (t *MockPanicTask) Stop() {
	// No operation needed for stopping.
}

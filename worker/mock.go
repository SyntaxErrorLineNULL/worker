package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// MockProcessingLongTaskCounter is a constant used to represent the value
// added to the contextDone counter when the processing is interrupted
// by a context cancellation. This value is used for testing purposes to
// verify that the task handles context cancellation correctly.
const MockProcessingLongTaskCounter = int32(100)

// MockProcessingLongTask is a mock implementation of the Processing interface designed to simulate long-running tasks.
// It is primarily used for testing purposes to validate how a task behaves when it takes a considerable amount of time to complete.
type MockProcessingLongTask struct {
	contextDone atomic.Int32
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
		// The context was canceled before the timeout elapses.
		// Increment the contextDone counter by MockProcessingLongTaskCounter to indicate interruption.
		m.contextDone.Add(MockProcessingLongTaskCounter)
		return
	case <-time.After(m.timeout):
		// Simulate long processing by blocking for the duration specified in m.timeout.
		// This is done using time.After to block the goroutine until the timeout has elapsed.
		fmt.Println("timeout")
	}
}

// ErrorHandler is a mock implementation of the error handling method.
// This method is a no-op (no operation) in this mock, as it is not used in the current test scenarios.
// In a real-world implementation, this method would handle any errors encountered during the processing.
func (m *MockProcessingLongTask) ErrorHandler(ctx context.Context, input interface{}) {
	// No operation in this mock implementation.
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

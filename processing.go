package worker

import "context"

// Processing defines an interface for processing tasks.
// It includes methods for performing the main processing logic and handling errors.
//
//go:generate mockery --name=Processing
type Processing interface {
	// Processing is responsible for executing the basic logic for processing the task.
	// It receives the task to be processed and the context.
	// The result of the execution can be obtained via the Result channel
	Processing(ctx context.Context, input interface{})
}

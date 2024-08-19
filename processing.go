package worker

import "context"

// Processing defines an interface for processing tasks.
// It includes methods for performing the main processing logic and handling errors.
//
//go:generate mockery --name=Processing
type Processing interface {
	Processing(ctx context.Context, input interface{}) (interface{}, error)

	ErrorHandler(ctx context.Context, input interface{}) (interface{}, error)
}

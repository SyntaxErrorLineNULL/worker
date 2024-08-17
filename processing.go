package worker

import "context"

type Processing interface {
	Processing(ctx context.Context, payload interface{})

	ErrorHandler(ctx context.Context, payload interface{})
}

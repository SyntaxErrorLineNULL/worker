package worker

import "context"

type Options struct {
	Context     context.Context
	Queue       chan Task
	WorkerCount int32
}

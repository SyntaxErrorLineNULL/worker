package worker

import (
	"context"
	"fmt"
	wr "github.com/SyntaxErrorLineNULL/worker"
	"github.com/davecgh/go-spew/spew"
	"runtime"
	"sync/atomic"
	"testing"
)

// BenchProcessingWithAtomic is a type that implements the Processing interface.
// It provides a concrete implementation for processing tasks in a benchmarking context.
type BenchProcessingWithAtomic struct {
	counter atomic.Int32
}

// Processing calculates the factorial of the given integer input using memoization.
// This method is intended to be used for benchmarking purposes, where it processes
// tasks by computing the factorial of an integer.
func (b *BenchProcessingWithAtomic) Processing(_ context.Context, _ interface{}) {
	b.counter.Add(1)
	return
}

func BenchmarkWorkerPool(b *testing.B) {
	// Define the number of workers in the pool.
	// This number determines how many concurrent worker goroutines will be used
	// to process tasks. The performance of the worker pool will be evaluated with this configuration.
	workerCount := int32(16)
	// Create a context to manage the lifecycle of the worker pool.
	// The context is used to control the cancellation and timeout of tasks within the pool.
	parentCtx := context.Background()
	// Create a buffered channel for task submission.
	// The buffer size is set to b.N, the number of iterations for the benchmark,
	// which ensures that tasks can be queued without blocking the worker pool.
	task := make(chan wr.Task)

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := NewWorkerPool(&wr.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, MaxRetryWorkerRestart: 3})

	// Start the worker pool in a separate Goroutine to allow it to operate asynchronously.
	// This enables the pool to begin its job processing and worker management in parallel.
	go pool.Run()

	// Add workers to the pool up to the defined worker count.
	// Each worker will be responsible for processing jobs from the pool.
	for w := int32(1); w <= workerCount; w++ {
		// Add a new worker to the worker pool.
		// The worker is created with a unique ID (in this case, hardcoded as 1) and a timeout of 3 seconds.
		// The logger is passed to the worker to handle logging within the worker's operations.
		// This operation should succeed because the pool's limit is set to accommodate this number of workers.
		err := pool.AddWorker(NewWorker(fmt.Sprintf("worker::%d", w)))

		// Check if there was an error adding the worker to the pool.
		// If an error occurred, the benchmark fails and halts execution.
		// This ensures that any issues with worker addition are caught and reported.
		if err != nil {
			b.Fatal("failed add new worker")
			return
		}
	}

	if workerCount != pool.RunningWorkers() {
		b.Fatal("failed running workers")
	}

	// Reset the benchmark timer to exclude setup time from the performance measurement.
	// This ensures that only the task processing time is measured.
	b.ResetTimer()

	// Create an instance of BenchProcessingWithAtomic, which implements the Processing interface.
	// This instance will handle the processing of factorial tasks during the benchmark.
	processing := &BenchProcessingWithAtomic{}

	// Iterate b.N times, where b.N is the number of iterations for the benchmark.
	// Each iteration represents a task to be processed by the worker pool.
	for i := 0; i < b.N; i++ {
		// Create a new task using the NewTask function.
		// This task is initialized with parameters such as ID (0), name ("test"),
		// the processing handler (processing), and an arbitrary value (i).
		// Note: The value used here (1) is a placeholder and might be replaced with
		// actual parameters depending on the implementation of NewTask.
		newTask := NewTask(0, "test", processing, i)
		// Associate the job with the context.
		// The context is used to manage the task's lifecycle, handle cancellations,
		// and control timeouts if necessary.
		_ = newTask.SetContext(parentCtx)

		// Submit the newly created task to the worker pool's task channel.
		// This enqueues the task for processing by the available workers.
		task <- newTask
	}

	spew.Dump(processing.counter.Load())

	// Force garbage collection to ensure accurate benchmark results.
	// This cleans up any memory used during the benchmark to avoid skewed results.
	defer runtime.GC()
}

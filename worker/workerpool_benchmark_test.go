package worker

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"

	wr "github.com/SyntaxErrorLineNULL/worker"
)

// mt is a mutex used to synchronize access to the factorialCache.
// This ensures that only one goroutine can read or write to the cache at a time,
// preventing race conditions and ensuring thread safety.
var mt sync.Mutex

// factorialCache stores previously computed factorial results to optimize performance
// by avoiding redundant calculations. The cache is initialized with the factorial of 0.
var factorialCache = map[int]int{
	0: 1,
}

// FactorialMemoized calculates the factorial of a given number n using memoization.
// It checks if the result is already available in the cache to avoid redundant computations.
// If the result is not in the cache, it recursively calculates the factorial,
// updates the cache, and then returns the result.
//
// The function uses a mutex (mt) to synchronize access to the factorialCache,
// ensuring that concurrent goroutines do not cause race conditions when reading from
// or writing to the cache.
func FactorialMemoized(n int) int {
	// Lock the mutex to ensure exclusive access to the factorialCache.
	mt.Lock()
	defer mt.Unlock()

	// Check if the result is already in the cache.
	if result, ok := factorialCache[n]; ok {
		return result
	}

	// If the result is not in the cache, calculate it recursively.
	result := n * FactorialMemoized(n-1)

	// Update the cache with the newly computed result.
	factorialCache[n] = result

	// Return the computed factorial result.
	return result
}

// BenchProcessing is a type that implements the Processing interface.
// It provides a concrete implementation for processing tasks in a benchmarking context.
type BenchProcessing struct{}

// Processing calculates the factorial of the given integer input using memoization.
// This method is intended to be used for benchmarking purposes, where it processes
// tasks by computing the factorial of an integer.
func (b *BenchProcessing) Processing(_ context.Context, input interface{}) {
	n := input.(int)
	_ = FactorialMemoized(n)
}

func BenchmarkWorkerPool(b *testing.B) {
	// BenchmarkWorkerPool-8   	  239079	      4339 ns/op 	no buffer chan
	// BenchmarkWorkerPool-8   	  491161	      3206 ns/op    26 worker
	// BenchmarkWorkerPool-8   	  391633	      2635 ns/op
	// BenchmarkWorkerPool-8   	  748088	      2658 ns/op	b.N chan buffer
	// BenchmarkWorkerPool-8   	  840723	      1478 ns/op	16 workers, chan buffer b.N

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
	task := make(chan wr.Task, b.N)

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := NewWorkerPool(&wr.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount})

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

	// Reset the benchmark timer to exclude setup time from the performance measurement.
	// This ensures that only the task processing time is measured.
	b.ResetTimer()

	// Create an instance of BenchProcessing, which implements the Processing interface.
	// This instance will handle the processing of factorial tasks during the benchmark.
	factorialProcessing := &BenchProcessing{}

	// Iterate b.N times, where b.N is the number of iterations for the benchmark.
	// Each iteration represents a task to be processed by the worker pool.
	for i := 0; i < b.N; i++ {
		// Create a new task using the NewTask function.
		// This task is initialized with parameters such as ID (0), name ("test"),
		// the processing handler (factorialProcessing), and an arbitrary value (1).
		// Note: The value used here (1) is a placeholder and might be replaced with
		// actual parameters depending on the implementation of NewTask.
		newTask := NewTask(0, "test", factorialProcessing, 1)
		// Associate the job with the context.
		// The context is used to manage the task's lifecycle, handle cancellations,
		// and control timeouts if necessary.
		_ = newTask.SetContext(parentCtx)

		// Submit the newly created task to the worker pool's task channel.
		// This enqueues the task for processing by the available workers.
		task <- newTask
	}

	// Force garbage collection to ensure accurate benchmark results.
	// This cleans up any memory used during the benchmark to avoid skewed results.
	defer runtime.GC()
}

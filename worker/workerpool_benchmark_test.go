//nolint:all
package worker

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/big"
	"math/rand"
	"runtime"
	"sync/atomic"
	"testing"

	wr "github.com/SyntaxErrorLineNULL/worker"

	"github.com/davecgh/go-spew/spew"
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
	// to currentProcess tasks. The performance of the worker pool will be evaluated with this configuration.
	workerCount := int32(16)
	// Create a context to manage the lifecycle of the worker pool.
	// The context is used to control the cancellation and timeout of tasks within the pool.
	parentCtx := context.Background()
	// Create a buffered channel for task submission.
	// The buffer size is set to b.N, the number of iterations for the benchmark,
	// which ensures that tasks can be queued without blocking the worker pool.
	task := make(chan wr.Task)

	// Initialize a logger for recording events during the benchmark.
	// This logger will help in capturing any logs generated during the benchmark test.
	logger := log.Default()

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := NewPool(&wr.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, MaxRetryWorkerRestart: 3, Logger: logger})

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
		err := pool.AddWorker(NewWorker(fmt.Sprintf("worker::%d", w), 0, logger))

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

// BenchmarkFactorialProcessing is a struct used for processing factorial calculations in a benchmark.
// It implements the logic needed to compute the factorial of an integer using a loop.
type BenchmarkFactorialProcessing struct{}

// NewFactorialProcessing creates a new instance of BenchmarkFactorialProcessing.
// This instance will be used in benchmarks to handle factorial calculations.
func NewFactorialProcessing() *BenchmarkFactorialProcessing {
	return &BenchmarkFactorialProcessing{}
}

// Processing calculates the factorial of the given integer input using memoization.
// This method is intended to be used for benchmarking purposes, where it processes
// tasks by computing the factorial of an integer.
func (b *BenchmarkFactorialProcessing) Processing(_ context.Context, input interface{}) {
	x := input.(int)
	result := big.NewInt(1)
	for i := 2; i <= x; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}
}

// Processing computes the factorial of an integer input using a loop-based approach.
// The method multiplies integers from 1 up to the given number and stores the result
// in a big integer to handle large values.
// This method is designed to simulate the computation of tasks in a benchmark environment.
func BenchmarkWorkerPoolFactorial(b *testing.B) {
	// Define the number of workers in the pool.
	// This number determines how many concurrent worker goroutines will be used
	// to currentProcess tasks. The performance of the worker pool will be evaluated with this configuration.
	workerCount := int32(16)
	// Create a context to manage the lifecycle of the worker pool.
	// The context is used to control the cancellation and timeout of tasks within the pool.
	parentCtx := context.Background()
	// Create a buffered channel for task submission.
	// The buffer size is set to b.N, the number of iterations for the benchmark,
	// which ensures that tasks can be queued without blocking the worker pool.
	task := make(chan wr.Task)

	// Initialize a logger for recording events during the benchmark.
	// This logger will help in capturing any logs generated during the benchmark test.
	logger := log.Default()

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := NewPool(&wr.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, MaxRetryWorkerRestart: 3, Logger: logger})

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
		err := pool.AddWorker(NewWorker(fmt.Sprintf("worker::%d", w), 0, logger))

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

	// Create an instance of BenchProcessingWithFactorial, which implements the Processing interface.
	// This instance will handle the processing of factorial tasks during the benchmark.
	processing := NewFactorialProcessing()

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

	// Force garbage collection to ensure accurate benchmark results.
	// This cleans up any memory used during the benchmark to avoid skewed results.
	defer runtime.GC()
}

type ImageProcessingStdLib struct {
	width  int
	height int
}

// NewImageProcessingStdLib creates a new instance of ImageProcessingStdLib with specified dimensions.
func NewImageProcessingStdLib(width, height int) *ImageProcessingStdLib {
	return &ImageProcessingStdLib{width: width, height: height}
}

// Processing inverts the colors of a generated image.
func (i *ImageProcessingStdLib) Processing(_ context.Context, _ interface{}) {
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, i.width, i.height))

	// Fill with random colors (simulating an image)
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			r := uint8(rand.Intn(256))
			g := uint8(rand.Intn(256))
			b := uint8(rand.Intn(256))
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Invert colors (white to black, etc.)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			original := img.At(x, y).(color.RGBA)
			inverted := color.RGBA{
				R: 255 - original.R,
				G: 255 - original.G,
				B: 255 - original.B,
				A: 255,
			}
			img.Set(x, y, inverted)
		}
	}

	// Simulate saving (in practice, write to a buffer or file)
	_ = img // Avoid unused variable warning
}

// BenchmarkWorkerPoolImageProcessingStdLib benchmarks the worker pool with image processing tasks using stdlib.
func BenchmarkWorkerPoolImageProcessingStdLib(b *testing.B) {
	// Define the number of workers in the pool.
	// This number determines how many concurrent worker goroutines will be used
	// to currentProcess tasks. The performance of the worker pool will be evaluated with this configuration.
	workerCount := int32(18)
	// Create a context to manage the lifecycle of the worker pool.
	// The context is used to control the cancellation and timeout of tasks within the pool.
	parentCtx := context.Background()
	// Create a buffered channel for task submission.
	// The buffer size is set to b.N, the number of iterations for the benchmark,
	// which ensures that tasks can be queued without blocking the worker pool.
	task := make(chan wr.Task, 1)

	// Initialize a logger for recording events during the benchmark.
	// This logger will help in capturing any logs generated during the benchmark test.
	logger := log.Default()

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := NewPool(&wr.Options{Context: parentCtx, Queue: task, WorkerCount: workerCount, MaxRetryWorkerRestart: 3, Logger: logger})

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
		err := pool.AddWorker(NewWorker(fmt.Sprintf("worker::%d", w), 0, logger))

		// Check if there was an error adding the worker to the pool.
		// If an error occurred, the benchmark fails and halts execution.
		// This ensures that any issues with worker addition are caught and reported.
		if err != nil {
			b.Fatal("failed add new worker")
			return
		}
	}

	// Verify the number of running workers.
	if workerCount != pool.RunningWorkers() {
		b.Fatalf("expected %d running workers, got %d", workerCount, pool.RunningWorkers())
	}

	// Create a processing instance for 1920x1080 images.
	processing := NewImageProcessingStdLib(420, 420)

	// Reset the timer to exclude setup time.
	b.ResetTimer()

	// Submit tasks in parallel.
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create a new task using the NewTask function.
			// This task is initialized with parameters such as ID (0), name ("test"),
			// the processing handler (processing), and an arbitrary value (i).
			// Note: The value used here (1) is a placeholder and might be replaced with
			// actual parameters depending on the implementation of NewTask.
			newTask := NewTask(0, "image-invert", processing, b.N)
			// Associate the job with the context.
			// The context is used to manage the task's lifecycle, handle cancellations,
			// and control timeouts if necessary.
			_ = newTask.SetContext(parentCtx)

			// Submit the newly created task to the worker pool's task channel.
			// This enqueues the task for processing by the available workers.
			task <- newTask
		}
	})

	// Force garbage collection to ensure accurate benchmark results.
	// This cleans up any memory used during the benchmark to avoid skewed results.
	defer runtime.GC()
}

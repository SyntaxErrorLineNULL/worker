# Golang worker pool lib

## worker-pool: Manage Concurrent Tasks with Ease in Go

worker-pool provides a well-structured and efficient library for managing concurrent tasks in Go. It offers a pool of worker goroutines that can be assigned tasks for parallel processing, improving performance and simplifying your application's workflow.

![Coverage](https://img.shields.io/badge/Coverage-84%25-brightgreen.svg)
![GitHub License](https://img.shields.io/github/license/SyntaxErrorLineNULL/worker)

## Install
```bash
go get github.com/SyntaxErrorLineNULL/worker
```

## Example:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	wr "github.com/SyntaxErrorLineNULL/worker"
	"github.com/SyntaxErrorLineNULL/worker/worker"
)

// Processing struct, which contains an atomic counter.
// The counter is used to safely update an integer value across multiple goroutines.
type Processing struct {
	counter atomic.Int32
}

// Processing method, which accepts a context and an input interface.
// This method processes the input data by casting it to an int32 and adding the value
// to the atomic counter. The use of atomic ensures thread-safe updates.
func (p *Processing) Processing(_ context.Context, input interface{}) {
	// Cast the input to an int32 value. This assumes that the input will always be an int32 type.
	// If the input is of a different type, a panic will occur, so care should be taken to ensure the correct type is passed.
	value := input.(int32)

	// Add the input value to the atomic counter. This ensures that the counter is updated in a thread-safe manner,
	// allowing multiple goroutines to safely increment the counter without causing race conditions.
	p.counter.Add(value)
}

func main() {
	// Create a context to manage the lifecycle of the worker pool.
	// The context is used to control the cancellation and timeout of tasks within the pool.
	parentCtx := context.Background()

	// Create a channel to hold tasks of type Task.
	// The channel is unbuffered, meaning that sends will block until a receiver is ready to receive the task.
	queue := make(chan wr.Task)

	// Define the number of workers in the pool.
	// This number determines how many concurrent worker goroutines will be used
	// to process tasks. The performance of the worker pool will be evaluated with this configuration.
	workerCount := int32(16)

	// Initialize a new worker pool with the specified context, task queue, and worker count.
	// The pool will manage the workers and distribute tasks to them for processing.
	pool := worker.NewWorkerPool(&wr.Options{Context: parentCtx, Queue: queue, WorkerCount: workerCount, MaxRetryWorkerRestart: 3})

	// Add workers to the pool up to the defined worker count.
	// Each worker will be responsible for processing jobs from the pool.
	for w := int32(1); w <= workerCount; w++ {
		// Add a new worker to the worker pool.
		// The worker is created with a unique ID (in this case, hardcoded as 1) and a timeout of 3 seconds.
		// The logger is passed to the worker to handle logging within the worker's operations.
		// This operation should succeed because the pool's limit is set to accommodate this number of workers.
		err := pool.AddWorker(worker.NewWorker(fmt.Sprintf("worker::%d", w)))
		// Check if the worker addition was successful.
		// If adding the worker fails for any reason, an error message is logged to indicate the failure.
		if err != nil {
			log.Println("Failed to add worker")
		}
	}

	// Create an instance of BenchProcessingWithAtomic, which implements the Processing interface.
	// This instance will handle the processing of factorial tasks during the benchmark.
	processing := &Processing{}

	taskCount := int32(1000)

	// Iterate b.N times, where b.N is the number of iterations for the benchmark.
	// Each iteration represents a task to be processed by the worker pool.
	for i := int32(0); i < taskCount; i++ {
		// Create a new task with timeout 0, name “test”, processing logic and current iteration value as input data.
		// The timeout value (in this case 0, which means without controlling the task execution time) is used to control the task processing time.
		// The 'i' value is passed as input data for task processing.
		newTask := worker.NewTask(0, "test", processing, i)
		// Attempt to add the new task to the task queue in the pool.
		if err := pool.AddTaskInQueue(newTask); err != nil {
			log.Println("Failed new task in queue")
		}
	}

	<-time.After(2 * time.Second)

	// Stop the worker pool.
	// The 'pool.Stop()' function call signals the pool to stop accepting new tasks and
	// gracefully shuts down all the workers after completing any ongoing tasks.
	pool.Stop()
	log.Println("worker pool is stop")
}
```

## SetDoneChannel

What is SetDoneChannel?

The **SetDoneChannel** method, as part of a Task interface, allows you to establish a channel that the task will use to signal its successful execution. This channel is typically closed once the task finishes its work. This mechanism enables other parts of your system to be notified of the task's completion and take appropriate actions.

### Example Usage:

Here's a code example demonstrating how to utilize **SetDoneChannel** within your application:

```go
  ...

  // Create a buffered channel with capacity 1 to signal task completion
  doneCh := make(chan struct{}, 1)

  // Instantiate a new task object
  newTask := worker.NewTask(0, "test", processing, int32(1))

  // Associate the done channel with the task
  err := newTask.SetDoneChannel(doneCh)
  if err != nil {
    fmt.Println("Error setting done channel:", err)
    return
  }

  // Add the task to your worker pool
  if err := pool.AddTaskInQueue(newTask); err != nil {
    fmt.Println("Error adding task to pool:", err)
    return
  }

  // Wait for completion or timeout
  select {
  case <-doneCh:
    fmt.Printf("counter: %d\n", processing.counter.Load())
    fmt.Println("Task completed successfully!")
  case <-time.After(2 * time.Second):
    fmt.Println("Timeout waiting for task to complete")
  }
```

## Pool interface

### AddWorker metod

#### Interface-Based Approach:

By accepting a Worker interface as an argument, the AddWorker method promotes flexibility and modularity. This enables you to implement different types of workers with varying behaviors and functionalities while maintaining a consistent interface for adding them to the pool. 

```go
// AddWorker registers a new worker to the pool. It initializes the worker with the pool's context,
// task queue, and error handling mechanisms. The worker begins processing tasks after being added.
// If the pool has been stopped or the worker cannot be added, an error is returned.
AddWorker(wr Worker) error
```

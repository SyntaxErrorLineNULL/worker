package worker

// Status represents the current state of a worker within a worker pool system.
// It is used to indicate and manage the different phases of a worker's lifecycle.
type Status int

// The possible values for WorkerStatus.
// These constants are used to manage and monitor the state of workers, allowing for efficient
// coordination and handling of workers within the pool.
const (
	// StatusWorkerIdle represents a state where the worker is waiting for a job to process.
	// In this state, the worker is ready and available to take on new tasks.
	StatusWorkerIdle Status = iota
	// StatusWorkerRunning represents a state where the worker is actively running a job.
	// This state indicates that the worker is engaged in processing a task and is not available for new jobs.
	StatusWorkerRunning
	// StatusWorkerStopped represents a state where the worker has stopped processing jobs.
	// This state indicates that the worker has been shut down or is no longer operational.
	StatusWorkerStopped
)

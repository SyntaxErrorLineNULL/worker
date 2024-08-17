package worker

// Error encapsulates an error along with the reference to the Worker instance
// where the error occurred. It is used to convey error information from a Worker
// to external components through the error channel.
type Error struct {
	Error    error  // The actual error that occurred.
	Instance Worker // Reference to the Worker instance where the error occurred.
}

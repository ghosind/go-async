package async

import "context"

// AsyncFn is the function to run, it needs to accept a `context.Context` object and return an
// error object.
type AsyncFn func(context.Context) error

// executeResult indicates the execution result whether the function returns an error or panic, and
// the index of the function in the parameters list.
type executeResult struct {
	// Error is the execution result of the function, it will be nil if the function does not return
	// an error and does not panic.
	Error error
	// Index is the index of the function in the parameters list.
	Index int
}

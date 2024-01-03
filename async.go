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

// getContext returns the specified non-nil context from the parameter, or creates and returns a
// new empty context.
func getContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}

	return context.Background()
}

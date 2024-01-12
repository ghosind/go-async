package async

import "errors"

var (
	// ErrContextCanceled to indicate the context was canceled or timed out.
	ErrContextCanceled error = errors.New("context canceled")
	// ErrInvalidConcurrency to indicate the number of the concurrency limitation is an invalid
	// value.
	ErrInvalidConcurrency error = errors.New("invalid concurrency")
	// ErrNotFunction indicates the value is not a function.
	ErrNotFunction error = errors.New("not function")
)

package async

import (
	"errors"
	"fmt"
)

var (
	// ErrContextCanceled to indicate the context was canceled or timed out.
	ErrContextCanceled error = errors.New("context canceled")
	// ErrInvalidConcurrency to indicate the number of the concurrency limitation is an invalid
	// value.
	ErrInvalidConcurrency error = errors.New("invalid concurrency")
	// ErrNotFunction indicates the value is not a function.
	ErrNotFunction error = errors.New("not function")
)

type ExecutionError interface {
	// Index returns the function's index in the parameters list that the function had returned an
	// error or panicked.
	Index() int
	// Err returns the original error that was returned or panicked by the function.
	Err() error
	// Error returns the execution error message.
	Error() string
}

type executionError struct {
	index int
	err   error
}

// Index returns the function's index in the parameters list that the function had returned an
// error or panicked.
func (e *executionError) Index() int {
	return e.index
}

// Err returns the original error that was returned or panicked by the function.
func (e *executionError) Err() error {
	return e.err
}

// Error returns the execution error message.
func (e *executionError) Error() string {
	return fmt.Sprintf("function %d error: %s", e.index, e.err)
}

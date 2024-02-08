package async

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrContextCanceled to indicate the context was canceled or timed out.
	ErrContextCanceled error = errors.New("context canceled")
	// ErrInvalidConcurrency to indicate the number of the concurrency limitation is an invalid
	// value.
	ErrInvalidConcurrency error = errors.New("invalid concurrency")
	// ErrNotFunction indicates the value is not a function.
	ErrNotFunction error = errors.New("not function")
	// ErrUnmatchedParam indicates the function's parameter list does not match to the list from the
	// caller.
	ErrUnmatchedParam error = errors.New("parameters are unmatched")
	// ErrInvalidTestFunc indicates the test function is invalid.
	ErrInvalidTestFunc error = errors.New("invalid test function")
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

// executionError is the error to represents the error of the function that is returned or
// panicked, and the index of the function in the parameters list.
type executionError struct {
	// index is the index of the function in the parameters list.
	index int
	// err is the error that the function returned or panicked.
	err error
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

// ExecutionErrors is an array of ExecutionError.
type ExecutionErrors []ExecutionError

// Error combines and returns all of the execution errors' message.
func (ee ExecutionErrors) Error() string {
	buf := bytes.NewBufferString("")

	for _, e := range ee {
		buf.WriteString(e.Error())
		buf.WriteByte('\n')
	}

	return strings.TrimSpace(buf.String())
}

// convertErrorListToExecutionErrors converts an array of the errors to the ExecutionErrors, it
// will set the index as the execution error's index. If the error in the list is nil, it will skip
// it and not add the error to the ExecutionErrors.
func convertErrorListToExecutionErrors(errs []error, num int) ExecutionErrors {
	if num == 0 {
		num = len(errs)
	}

	ee := make(ExecutionErrors, 0, num)

	for i, e := range errs {
		if e == nil {
			continue
		}

		ee = append(ee, &executionError{
			index: i,
			err:   e,
		})
	}

	if len(ee) == 0 {
		return nil
	}

	return ee
}

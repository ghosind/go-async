package async

import (
	"errors"
	"testing"

	"github.com/ghosind/go-assert"
)

func TestExecutionError(t *testing.T) {
	a := assert.New(t)

	innerErr := errors.New("inner error")
	err := &executionError{
		index: 0,
		err:   innerErr,
	}

	a.EqualNow(err.Error(), "function 0 error: inner error")
	a.EqualNow(err.Index(), 0)
	a.EqualNow(err.Err(), innerErr)
}

func TestExecutionErrors(t *testing.T) {
	a := assert.New(t)

	var ee ExecutionErrors = []ExecutionError{
		&executionError{
			index: 0,
			err:   errors.New("inner error 1"),
		},
		&executionError{
			index: 1,
			err:   errors.New("inner error 2"),
		},
	}

	a.EqualNow(ee.Error(), `function 0 error: inner error 1
function 1 error: inner error 2`)
}

func TestConvertErrorListToExecutionErrors(t *testing.T) {
	a := assert.New(t)

	a.NilNow(convertErrorListToExecutionErrors(nil, 0))
	a.NilNow(convertErrorListToExecutionErrors(nil, 1))
	a.NilNow(convertErrorListToExecutionErrors([]error{}, 0))
	a.NilNow(convertErrorListToExecutionErrors([]error{}, 1))
	a.NilNow(convertErrorListToExecutionErrors([]error{nil}, 0))
	a.NilNow(convertErrorListToExecutionErrors([]error{nil}, 1))
	a.EqualNow(
		convertErrorListToExecutionErrors([]error{errors.New("expected error")}, 1).Error(),
		"function 0 error: expected error")
	a.EqualNow(
		convertErrorListToExecutionErrors([]error{errors.New("expected error"), nil}, 0).Error(),
		"function 0 error: expected error")
	a.EqualNow(
		convertErrorListToExecutionErrors([]error{
			errors.New("expected error 1"),
			errors.New("expected error 2"),
		}, 1).Error(),
		`function 0 error: expected error 1
function 1 error: expected error 2`)
}

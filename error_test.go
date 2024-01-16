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

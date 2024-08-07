//go:build go1.20

package async

import (
	"errors"
	"testing"

	"github.com/ghosind/go-assert"
)

func TestUnwrapExecutionsError(t *testing.T) {
	a := assert.New(t)

	innerErr := errors.New("expected error")
	err := &executionError{
		err:   innerErr,
		index: 0,
	}

	errs := ExecutionErrors{err}
	a.IsErrorNow(errs, err)
	a.IsErrorNow(errs, innerErr)
}

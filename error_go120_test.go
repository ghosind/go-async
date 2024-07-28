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
	a.TrueNow(errors.Is(errs, err))
	a.TrueNow(errors.Is(errs, innerErr))
}

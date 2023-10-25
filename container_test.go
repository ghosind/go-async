package async

import (
	"errors"
	"testing"

	"github.com/ghosind/go-assert"
)

func TestExecutionContainer(t *testing.T) {
	a := assert.New(t)

	testExecutionContainer(a, func() error {
		return nil
	}, false)
	testExecutionContainer(a, func() error {
		return errors.New("expected error")
	}, true, "expected error")
	testExecutionContainer(a, func() error {
		panic("expected panic")
	}, true, "expected panic")
	testExecutionContainer(a, func() error {
		panic(errors.New("expected panic error"))
	}, true, "expected panic error")
	testExecutionContainer(a, func() error {
		panic(123)
	}, true, "123")
}

func testExecutionContainer(a *assert.Assertion, fn func() error, hasError bool, message ...string) {
	err := executionContainer(fn)

	if !hasError {
		a.NilNow(err)
	} else {
		a.NotNilNow(err)

		if len(message) > 0 {
			a.Equal(err.Error(), message[0])
		}
	}
}

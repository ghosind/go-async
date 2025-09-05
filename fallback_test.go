package async_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestFallback(t *testing.T) {
	a := assert.New(t)
	runCount := 0

	err := async.Fallback(
		func() error {
			runCount++
			return errors.New("first error")
		},
		func() error {
			runCount++
			return nil
		},
		func() error {
			runCount++
			return errors.New("last error")
		},
	)
	a.NilNow(err)
	a.EqualNow(runCount, 2)
}

func TestFallbackAllFailed(t *testing.T) {
	a := assert.New(t)
	runCount := 0
	finalErr := errors.New("third error")

	err := async.Fallback(
		func() error {
			runCount++
			return errors.New("first error")
		},
		func() error {
			runCount++
			return errors.New("second error")
		},
		func() error {
			runCount++
			return finalErr
		},
	)
	a.IsErrorNow(err, finalErr)
	a.EqualNow(runCount, 3)
}

func TestFallbackWithContext(t *testing.T) {
	a := assert.New(t)
	runCount := 0

	err := async.FallbackWithContext(
		context.Background(),
		func() error {
			runCount++
			return errors.New("first error")
		},
		func() error {
			runCount++
			return nil
		},
		func() error {
			runCount++
			return errors.New("last error")
		},
	)
	a.NilNow(err)
	a.EqualNow(runCount, 2)
}

func ExampleFallback() {
	err := async.Fallback(func() error {
		return errors.New("first error")
	}, func() error {
		return errors.New("second error")
	}, func() error {
		return nil
	}, func() error {
		return errors.New("third error")
	})
	if err != nil {
		// handle the error
	}
	// err: <nil>
}

package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestAllWithoutFuncs(t *testing.T) {
	a := assert.New(t)

	index, err := All()
	a.NilNow(err)
	a.EqualNow(index, -1)
}

func TestAllSuccess(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	index, err := All(funcs...)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestAllFailure(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			if n == 2 {
				return errors.New("n = 2")
			}
			data[n] = true
			return nil
		})
	}

	index, err := All(funcs...)
	a.NotNilNow(err)
	a.EqualNow(index, 2)
	a.EqualNow(err.Error(), "n = 2")
	a.EqualNow(data, []bool{true, true, false, false, false})
}

func TestAllWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	index, err := AllWithContext(nil, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	a.NilNow(err)
	a.EqualNow(index, -1)
}

func TestAllWithTimeoutContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	index, err := AllWithContext(ctx, funcs...)
	a.NotNilNow(err)
	a.EqualNow(index, -1)
	a.Equal(err.Error(), "context canceled")
	a.EqualNow(data, []bool{true, true, false, false, false})
}

func TestAllCompletedWithoutFuncs(t *testing.T) {
	a := assert.New(t)

	errs, hasError := AllCompleted()
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{})
}

func TestAllCompletedSuccess(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	errs, hasError := AllCompleted(funcs...)
	a.NotTrueNow(hasError)
	a.EqualNow(data, []bool{true, true, true, true, true})
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
}

func TestAllCompletedPartialFailure(t *testing.T) {
	a := assert.New(t)

	errNIs2 := errors.New("n = 2")

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			if n == 2 {
				return errNIs2
			}
			data[n] = true
			return nil
		})
	}

	errs, hasError := AllCompleted(funcs...)
	a.TrueNow(hasError)
	a.EqualNow(data, []bool{true, true, false, true, true})
	a.EqualNow(errs, []error{nil, nil, errNIs2, nil, nil})
}

func TestAllCompletedWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	errs, hasError := AllCompletedWithContext(nil, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil})
}

func TestAllCompletedWithTimeoutContext(t *testing.T) {
	a := assert.New(t)

	errTimeout := errors.New("timeout")

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			select {
			case <-ctx.Done():
				return errTimeout
			default:
				data[n] = true
				return nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	errs, hasError := AllCompletedWithContext(ctx, funcs...)
	a.TrueNow(hasError)
	a.EqualNow(data, []bool{true, true, false, false, false})
	a.EqualNow(errs, []error{nil, nil, errTimeout, errTimeout, errTimeout})
}

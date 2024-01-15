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

	out, index, err := All()
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.NilNow(out)
}

func TestAllSuccess(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return n, nil
		})
	}

	out, index, err := All(funcs...)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.EqualNow(data, []bool{true, true, true, true, true})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestAllFailure(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("n = 2")

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			if n == 2 {
				return expectedErr
			}
			data[n] = true
			return nil
		})
	}

	out, index, err := All(funcs...)
	a.NotNilNow(err)
	a.EqualNow(index, 2)
	a.EqualNow(err, expectedErr)
	a.EqualNow(data, []bool{true, true, false, false, false})
	a.EqualNow(out, [][]any{{nil}, {nil}, {expectedErr}, nil, nil})
}

func TestAllWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	out, index, err := AllWithContext(nil, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.EqualNow(out, [][]any{{nil}})
}

func TestAllWithTimeoutContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
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

	out, index, err := AllWithContext(ctx, funcs...)
	a.NotNilNow(err)
	a.EqualNow(index, -1)
	a.TrueNow(errors.Is(err, ErrContextCanceled))
	a.EqualNow(data, []bool{true, true, false, false, false})
	a.EqualNow(out, [][]any{{nil}, {nil}, nil, nil, nil})
}

func BenchmarkAll(b *testing.B) {
	tasks := make([]AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	All(tasks...)
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
	funcs := make([]AsyncFn, 0, 5)
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
	funcs := make([]AsyncFn, 0, 5)
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
	funcs := make([]AsyncFn, 0, 5)
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

func BenchmarkAllCompleted(b *testing.B) {
	tasks := make([]AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	AllCompleted(tasks...)
}

package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestParallel(t *testing.T) {
	a := assert.New(t)

	index, err := Parallel(0)
	a.NilNow(err)
	a.EqualNow(index, -1)

	a.PanicNow(func() {
		Parallel(-1)
	})
}

func TestParallelWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	start := time.Now()
	index, err := Parallel(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.TrueNow(dur-100*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
}

func TestParallelWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	start := time.Now()
	index, err := Parallel(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.TrueNow(dur-300*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
}

func TestParallelWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			if n == 2 {
				return expectedErr
			}
			return nil
		})
	}

	index, err := Parallel(2, funcs...)
	a.EqualNow(err, expectedErr)
	a.EqualNow(index, 2)
}

func TestParallelWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return nil
		})
	}

	index, err := ParallelWithContext(context.Background(), 2, funcs...)
	a.NilNow(err)
	a.EqualNow(index, -1)

	finishedNum := 0
	for _, v := range res {
		if v {
			finishedNum++
		}
	}
	a.EqualNow(finishedNum, 5)
}

func TestParallelWithTimedOutContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return nil
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	index, err := ParallelWithContext(ctx, 2, funcs...)
	a.TrueNow(errors.Is(err, ErrContextCanceled))
	a.EqualNow(index, -1)

	finishedNum := 0
	for _, v := range res {
		if v {
			finishedNum++
		}
	}
	a.EqualNow(finishedNum, 2)
}

func TestParallelCompleted(t *testing.T) {
	a := assert.New(t)

	errs, hasError := ParallelCompleted(0)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{})

	a.PanicNow(func() {
		ParallelCompleted(-1)
	})
}

func TestParallelCompletedWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	start := time.Now()
	errs, hasError := ParallelCompleted(0, funcs...)
	dur := time.Since(start)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
	a.TrueNow(dur-100*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
}

func TestParallelCompletedWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	start := time.Now()
	errs, hasError := ParallelCompleted(2, funcs...)
	dur := time.Since(start)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
	a.TrueNow(dur-300*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
}

func TestParallelCompletedWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			if n == 2 {
				time.Sleep(50 * time.Millisecond)
				return expectedErr
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})
	}

	errs, hasError := ParallelCompleted(0, funcs...)
	a.TrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, expectedErr, nil, nil})
}

func TestParallelCompletedWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return nil
		})
	}

	errs, hasError := ParallelCompletedWithContext(context.Background(), 2, funcs...)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})

	finishedNum := 0
	for _, v := range res {
		if v {
			finishedNum++
		}
	}
	a.EqualNow(finishedNum, 5)
}

func TestParallelCompletedWithTimedOutContext(t *testing.T) {
	a := assert.New(t)

	timeoutErr := errors.New("timed out")

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			timer := time.NewTimer(100 * time.Millisecond)

			select {
			case <-ctx.Done():
				return timeoutErr
			case <-timer.C:
				res[n] = true
				return nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	errs, hasError := ParallelCompletedWithContext(ctx, 2, funcs...)
	a.TrueNow(hasError)

	numErrors := 0
	for _, e := range errs {
		if errors.Is(e, timeoutErr) {
			numErrors++
		}
	}
	a.EqualNow(numErrors, 3)

	finishedNum := 0
	for _, v := range res {
		if v {
			finishedNum++
		}
	}
	a.EqualNow(finishedNum, 2)
}

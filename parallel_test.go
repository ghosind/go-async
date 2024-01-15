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

	out, index, err := Parallel(0)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.EqualNow(out, [][]any{})

	a.PanicNow(func() {
		Parallel(-1)
	})
}

func TestParallelWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, index, err := Parallel(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.TrueNow(dur-100*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, index, err := Parallel(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.TrueNow(dur-300*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			if n == 2 {
				time.Sleep(50 * time.Millisecond)
				return n, expectedErr
			} else {
				time.Sleep(100 * time.Millisecond)
				return n, nil
			}
		})
	}

	start := time.Now()
	out, index, err := Parallel(2, funcs...)
	dur := time.Since(start)
	a.EqualNow(err, expectedErr)
	a.EqualNow(index, 2)
	a.TrueNow(dur-150*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, expectedErr}, nil, nil})
}

func TestParallelWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return n, nil
		})
	}

	out, index, err := ParallelWithContext(context.Background(), 2, funcs...)
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})

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
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return n, nil
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	out, index, err := ParallelWithContext(ctx, 2, funcs...)
	a.TrueNow(errors.Is(err, ErrContextCanceled))
	a.EqualNow(index, -1)
	a.EqualNow(res, []bool{true, true, false, false, false})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, nil, nil, nil})
}

func BenchmarkParallel(b *testing.B) {
	tasks := make([]AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	Parallel(5, tasks...)
}

func TestParallelCompleted(t *testing.T) {
	a := assert.New(t)

	out, errs, hasError := ParallelCompleted(0)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{})
	a.EqualNow(out, [][]any{})

	a.PanicNow(func() {
		ParallelCompleted(-1)
	})
}

func TestParallelCompletedWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, errs, hasError := ParallelCompleted(0, funcs...)
	dur := time.Since(start)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
	a.TrueNow(dur-100*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelCompletedWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, errs, hasError := ParallelCompleted(2, funcs...)
	dur := time.Since(start)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
	a.TrueNow(dur-300*time.Millisecond < 30*time.Millisecond) // allow 30ms deviation
}

func TestParallelCompletedWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			if n == 2 {
				time.Sleep(50 * time.Millisecond)
				return n, expectedErr
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			return n, nil
		})
	}

	out, errs, hasError := ParallelCompleted(0, funcs...)
	a.TrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, expectedErr, nil, nil})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, expectedErr}, {3, nil}, {4, nil}})
}

func TestParallelCompletedWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return n, nil
		})
	}

	out, errs, hasError := ParallelCompletedWithContext(context.Background(), 2, funcs...)
	a.NotTrueNow(hasError)
	a.EqualNow(errs, []error{nil, nil, nil, nil, nil})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})

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
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			timer := time.NewTimer(100 * time.Millisecond)

			select {
			case <-ctx.Done():
				return n, timeoutErr
			case <-timer.C:
				res[n] = true
				return n, nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	_, errs, hasError := ParallelCompletedWithContext(ctx, 2, funcs...)
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

func BenchmarkParallelCompleted(b *testing.B) {
	tasks := make([]AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	ParallelCompleted(5, tasks...)
}

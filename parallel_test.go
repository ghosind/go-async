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

	out, err := Parallel(0)
	a.NilNow(err)
	a.EqualNow(out, [][]any{})

	a.PanicOfNow(func() {
		Parallel(-1)
	}, ErrInvalidConcurrency)
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
	out, err := Parallel(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 130*time.Millisecond) // allow 30ms deviation
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
	out, err := Parallel(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 300*time.Millisecond)
	a.LteNow(dur, 350*time.Millisecond) // allow 50ms deviation
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
	out, err := Parallel(2, funcs...)
	dur := time.Since(start)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), "function 2 error: expected error")
	a.GteNow(dur, 150*time.Millisecond)
	a.LteNow(dur, 180*time.Millisecond) // allow 30ms deviation
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

	out, err := ParallelWithContext(context.Background(), 2, funcs...)
	a.NilNow(err)
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

	out, err := ParallelWithContext(ctx, 2, funcs...)
	a.NotNilNow(err)
	a.TrueNow(errors.Is(err, ErrContextCanceled))
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

	out, err := ParallelCompleted(0)
	a.NilNow(err)
	a.EqualNow(out, [][]any{})

	a.PanicOfNow(func() {
		ParallelCompleted(-1)
	}, ErrInvalidConcurrency)
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
	out, err := ParallelCompleted(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 130*time.Millisecond) // allow 30ms deviation
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
	out, err := ParallelCompleted(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
	a.GteNow(dur, 300*time.Millisecond)
	a.LteNow(dur, 350*time.Millisecond) // allow 50ms deviation
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

	out, err := ParallelCompleted(0, funcs...)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), "function 2 error: expected error")
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

	out, err := ParallelCompletedWithContext(context.Background(), 2, funcs...)
	a.NilNow(err)
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

	out, err := ParallelCompletedWithContext(ctx, 2, funcs...)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), `function 2 error: timed out
function 3 error: timed out
function 4 error: timed out`)
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, timeoutErr}, {3, timeoutErr}, {4, timeoutErr}})

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

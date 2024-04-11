package async_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestParallel(t *testing.T) {
	a := assert.New(t)

	out, err := async.Parallel(0)
	a.NilNow(err)
	a.EqualNow(out, [][]any{})

	a.PanicOfNow(func() {
		async.Parallel(-1)
	}, async.ErrInvalidConcurrency)
}

func TestParallelWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, err := async.Parallel(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 130*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, err := async.Parallel(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 300*time.Millisecond)
	a.LteNow(dur, 350*time.Millisecond) // allow 50ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]async.AsyncFn, 0, 5)
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
	out, err := async.Parallel(2, funcs...)
	dur := time.Since(start)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), "function 2 error: expected error")
	a.GteNow(dur, 150*time.Millisecond)
	a.LteNow(dur, 180*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, expectedErr}, nil, nil})
}

func TestParallelWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return n, nil
		})
	}

	out, err := async.ParallelWithContext(context.Background(), 2, funcs...)
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

	funcs := make([]async.AsyncFn, 0, 5)
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

	out, err := async.ParallelWithContext(ctx, 2, funcs...)
	a.NotNilNow(err)
	a.TrueNow(errors.Is(err, async.ErrContextCanceled))
	a.EqualNow(res, []bool{true, true, false, false, false})
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, nil, nil, nil})
}

func BenchmarkParallel(b *testing.B) {
	tasks := make([]async.AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	async.Parallel(5, tasks...)
}

func ExampleParallel() {
	out, err := async.Parallel(2, func() int {
		time.Sleep(50 * time.Millisecond)
		return 1
	}, func() int {
		time.Sleep(50 * time.Millisecond)
		return 2
	}, func() int {
		time.Sleep(50 * time.Millisecond)
		return 3
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// [[1] [2] [3]]
	// <nil>
}

func TestParallelCompleted(t *testing.T) {
	a := assert.New(t)

	out, err := async.ParallelCompleted(0)
	a.NilNow(err)
	a.EqualNow(out, [][]any{})

	a.PanicOfNow(func() {
		async.ParallelCompleted(-1)
	}, async.ErrInvalidConcurrency)
}

func TestParallelCompletedWithoutConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, err := async.ParallelCompleted(0, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 130*time.Millisecond) // allow 30ms deviation
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
}

func TestParallelCompletedWithConcurrencyLimit(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n, nil
		})
	}

	start := time.Now()
	out, err := async.ParallelCompleted(2, funcs...)
	dur := time.Since(start)
	a.NilNow(err)
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, nil}, {3, nil}, {4, nil}})
	a.GteNow(dur, 300*time.Millisecond)
	a.LteNow(dur, 350*time.Millisecond) // allow 50ms deviation
}

func TestParallelCompletedWithFailedTask(t *testing.T) {
	a := assert.New(t)

	expectedErr := errors.New("expected error")

	funcs := make([]async.AsyncFn, 0, 5)
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

	out, err := async.ParallelCompleted(0, funcs...)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), "function 2 error: expected error")
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {2, expectedErr}, {3, nil}, {4, nil}})
}

func TestParallelCompletedWithContext(t *testing.T) {
	a := assert.New(t)

	funcs := make([]async.AsyncFn, 0, 5)
	res := make([]bool, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			res[n] = true
			return n, nil
		})
	}

	out, err := async.ParallelCompletedWithContext(context.Background(), 2, funcs...)
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

	funcs := make([]async.AsyncFn, 0, 5)
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

	out, err := async.ParallelCompletedWithContext(ctx, 2, funcs...)
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
	tasks := make([]async.AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	async.ParallelCompleted(5, tasks...)
}

func ExampleParallelCompleted() {
	out, err := async.ParallelCompleted(2, func() int {
		time.Sleep(50 * time.Millisecond)
		return 1
	}, func() error {
		time.Sleep(50 * time.Millisecond)
		return errors.New("expected error")
	}, func() int {
		time.Sleep(50 * time.Millisecond)
		return 3
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// [[1] [expected error] [3]]
	// function 1 error: expected error
}

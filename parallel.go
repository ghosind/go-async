package async

import (
	"context"
	"sync"
	"sync/atomic"
)

// Parallel runs the functions asynchronously with the specified concurrency limitation. It will
// send a cancel sign to context and terminate immediately if any function returns an error or
// panic, and also returns an execution error to indicate the error.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
//
//	// Run 2 functions asynchronously at the time.
//	out, err := Parallel(2, func(ctx context.Context) (int, error) {
//	  // Do something
//	  return 1, nil
//	}, func(ctx context.Context) (string, error) {
//	  // Do something
//	  return "hello", nil
//	}, func(ctx context.Context) error {
//	// Do something
//	  return nil
//	} /* , ... */)
//	// out: [][]any{{1, <nil>}, {"hello", <nil>}, {<nil>}}
//	// err: <nil>
func Parallel(concurrency int, funcs ...AsyncFn) ([][]any, error) {
	return parallel(context.Background(), concurrency, funcs...)
}

// ParallelWithContext runs the functions asynchronously with the specified concurrency limitation.
// It will send a cancel sign to context and terminate immediately if any function returns an error
// or panic, and also returns an execution error to indicate the error. If the context was canceled
// or timed out before all functions finished executing, it will send a cancel sign to all
// uncompleted functions, and return a context canceled error.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func ParallelWithContext(
	ctx context.Context,
	concurrency int,
	funcs ...AsyncFn,
) ([][]any, error) {
	return parallel(ctx, concurrency, funcs...)
}

// parallel runs the functions asynchronously with the specified concurrency.
func parallel(parent context.Context, concurrency int, funcs ...AsyncFn) ([][]any, error) {
	paralleler := new(Paralleler).
		WithContext(parent).
		WithConcurrency(concurrency)

	paralleler.Add(funcs...)

	return paralleler.Run()
}

// ParallelCompleted runs the functions asynchronously with the specified concurrency limitation.
// It returns an error array and a boolean value to indicate whether any function panics or returns
// an error, and you can get the error details from the error array by the indices of the functions
// in the parameter list. It will return until all of the functions are finished.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func ParallelCompleted(concurrency int, funcs ...AsyncFn) ([][]any, error) {
	return parallelCompleted(context.Background(), concurrency, funcs...)
}

// ParallelCompletedWithContext runs the functions asynchronously with the specified concurrency
// limitation and the context. It returns an error array and a boolean value to indicate whether
// any function panics or returns an error, and you can get the error details from the error array
// by the indices of the functions in the parameter list. It will return until all of the functions
// are finished.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func ParallelCompletedWithContext(
	ctx context.Context,
	concurrency int,
	funcs ...AsyncFn,
) ([][]any, error) {
	return parallelCompleted(ctx, concurrency, funcs...)
}

// parallelCompleted runs the functions asynchronously with the specified concurrency until all of
// the functions are finished.
func parallelCompleted(
	parent context.Context,
	concurrency int,
	funcs ...AsyncFn,
) ([][]any, error) {
	// the number of concurrency should be 0 (no limitation) or greater than 0.
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}
	validateAsyncFuncs(funcs...)

	out := make([][]any, len(funcs))

	if len(funcs) == 0 {
		return out, nil
	}

	ctx := getContext(parent)
	errs := make([]error, len(funcs))
	errNum := atomic.Int32{}

	wg := sync.WaitGroup{}
	wg.Add(len(funcs))

	var conch chan empty // channel for concurrency limit
	// no concurrency limitation if the value of the number is 0
	if concurrency > 0 {
		conch = make(chan empty, concurrency)
	}

	for i := 0; i < len(funcs); i++ {
		if conch != nil {
			conch <- empty{}
		}

		go func(n int) {
			defer wg.Done()

			fn := funcs[n]
			childCtx, childCanFunc := context.WithCancel(ctx)
			defer childCanFunc()

			ret, err := invokeAsyncFn(fn, childCtx, nil)
			if err != nil {
				errs[n] = err
				errNum.Add(1)
			}
			out[n] = ret

			if conch != nil {
				<-conch
			}
		}(i)
	}

	wg.Wait()
	if errNum.Load() == 0 {
		return out, nil
	}

	err := convertErrorListToExecutionErrors(errs, int(errNum.Load()))

	return out, err
}

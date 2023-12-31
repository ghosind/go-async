package async

import (
	"context"
	"sync"

	"github.com/ghosind/utils"
)

// Parallel runs the functions asynchronously with the specified concurrency limitation. It will
// send a cancel sign to context and terminate immediately if any function returns an error or
// panic, and also returns the index of the error function in the parameters list.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func Parallel(concurrency int, funcs ...AsyncFn) (int, error) {
	return parallel(context.Background(), concurrency, funcs...)
}

// ParallelWithContext runs the functions asynchronously with the specified concurrency limitation.
// It will send a cancel sign to context and terminate immediately if any function returns an error
// or panic, and also returns the index of the error function in the parameters list. If the
// context was canceled or timed out before all functions finished executing, it will send a cancel
// sign to all uncompleted functions, and return the value of the index as -1 with a `context
// canceled` error.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func ParallelWithContext(ctx context.Context, concurrency int, funcs ...AsyncFn) (int, error) {
	return parallel(ctx, concurrency, funcs...)
}

// parallel runs the functions asynchronously with the specified concurrency.
func parallel(parent context.Context, concurrency int, funcs ...AsyncFn) (int, error) {
	// the number of concurrency should be 0 (no limitation) or greater than 0.
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}

	if len(funcs) == 0 {
		return -1, nil
	}

	parent = getContext(parent)
	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	ch := make(chan executeResult, len(funcs)) // channel for result
	var conch chan empty                       // channel for concurrency limit

	// no concurrency limitation if the value of the number is 0
	if concurrency > 0 {
		conch = make(chan empty, concurrency)
	}

	for i := 0; i < len(funcs); i++ {
		go runTaskInParallel(ctx, i, funcs[i], conch, ch)
	}

	finished := 0
	for finished < len(funcs) {
		select {
		case <-parent.Done():
			return -1, ErrContextCanceled
		case ret := <-ch:
			if ret.Error != nil {
				return ret.Index, ret.Error
			}
			finished++
		}
	}

	return -1, nil
}

// runTaskInParallel runs the specified function for Parallel / ParallelWithContext.
func runTaskInParallel(
	ctx context.Context,
	n int,
	fn AsyncFn,
	conch chan empty,
	ch chan executeResult,
) {
	if conch != nil {
		conch <- empty{}
	}

	childCtx, childCanFunc := context.WithCancel(ctx)
	defer childCanFunc()

	err := utils.Try(func() error {
		return fn(childCtx)
	})

	if conch != nil {
		<-conch
	}

	select {
	case <-ctx.Done():
		return
	default:
		ch <- executeResult{
			Index: n,
			Error: err,
		}
	}
}

// ParallelCompleted runs the functions asynchronously with the specified concurrency limitation.
// It returns an error array and a boolean value to indicate whether any function panics or returns
// an error, and you can get the error details from the error array by the indices of the functions
// in the parameter list. It will return until all of the functions are finished.
//
// The number of concurrency must be greater than or equal to 0, and it means no concurrency
// limitation if the number is 0.
func ParallelCompleted(concurrency int, funcs ...AsyncFn) ([]error, bool) {
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
) ([]error, bool) {
	return parallelCompleted(ctx, concurrency, funcs...)
}

// parallelCompleted runs the functions asynchronously with the specified concurrency until all of
// the functions are finished.
func parallelCompleted(parent context.Context, concurrency int, funcs ...AsyncFn) ([]error, bool) {
	// the number of concurrency should be 0 (no limitation) or greater than 0.
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}

	errs := make([]error, len(funcs))
	hasError := false

	if len(funcs) == 0 {
		return errs, hasError
	}

	ctx := getContext(parent)

	wg := sync.WaitGroup{}
	wg.Add(len(funcs))

	var conch chan empty // channel for concurrency limit
	// no concurrency limitation if the value of the number is 0
	if concurrency > 0 {
		conch = make(chan empty, concurrency)
	}

	for i := 0; i < len(funcs); i++ {
		go func(n int) {
			defer wg.Done()

			if conch != nil {
				conch <- empty{}
			}

			fn := funcs[n]
			childCtx, childCanFunc := context.WithCancel(ctx)
			defer childCanFunc()

			if err := utils.Try(func() error {
				return fn(childCtx)
			}); err != nil {
				errs[n] = err
				hasError = true
			}

			if conch != nil {
				<-conch
			}
		}(i)
	}

	wg.Wait()

	return errs, hasError
}

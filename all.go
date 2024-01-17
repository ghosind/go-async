package async

import (
	"context"
	"sync"
	"sync/atomic"
)

// All executes the functions asynchronously until all functions have been finished. If some
// function returns an error or panic, it will immediately return an execution error, and send a
// cancel signal to all other functions by context.
//
// The index of the function will be -1 if all functions have been completed without error or
// panic.
func All(funcs ...AsyncFn) ([][]any, error) {
	return all(context.Background(), funcs...)
}

// AllWithContext executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout). If some function returns an error or panic, it will
// immediately return an execution error and send a cancel signal to all other functions by
// context.
//
// The index of the function will be -1 if all functions have been completed without error or
// panic, or the context has been canceled (or timeout) before all functions finished.
func AllWithContext(ctx context.Context, funcs ...AsyncFn) ([][]any, error) {
	return all(ctx, funcs...)
}

// all executes the functions asynchronously until all functions have been finished, or the context
// is done (canceled or timeout).
func all(parent context.Context, funcs ...AsyncFn) ([][]any, error) {
	if len(funcs) == 0 {
		return nil, nil
	}
	validateAsyncFuncs(funcs...)

	parent = getContext(parent)

	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	ch := make(chan executeResult, len(funcs))
	defer close(ch)

	for i := 0; i < len(funcs); i++ {
		go runTaskInAll(ctx, i, funcs[i], ch)
	}

	finished := 0
	out := make([][]any, len(funcs))
	for finished < len(funcs) {
		select {
		case <-parent.Done():
			return out, ErrContextCanceled
		case ret := <-ch:
			out[ret.Index] = ret.Out
			if ret.Error != nil {
				return out, &executionError{
					index: ret.Index,
					err:   ret.Error,
				}
			}
			finished++
		}
	}

	return out, nil
}

// runTaskInAll runs the specified function for All / AllWithContext.
func runTaskInAll(ctx context.Context, n int, fn AsyncFn, ch chan<- executeResult) {
	childCtx, childCanFunc := context.WithCancel(ctx)
	defer childCanFunc()

	ret, err := invokeAsyncFn(fn, childCtx, nil)

	select {
	case <-ctx.Done():
		return
	default:
		ch <- executeResult{
			Error: err,
			Index: n,
			Out:   ret,
		}
	}
}

// AllCompleted executes the functions asynchronously until all functions have been finished. It
// will return an error slice that is ordered by the functions order, and a boolean value to
// indicate whether any functions return an error or panic.
func AllCompleted(funcs ...AsyncFn) ([][]any, error) {
	return allCompleted(context.Background(), funcs...)
}

// AllCompletedWithContext executes the functions asynchronously until all functions have been
// finished, or the context is done (canceled or timeout). It will return an error slice that is
// ordered by the functions order, and a boolean value to indicate whether any functions return an
// error or panic.
func AllCompletedWithContext(
	ctx context.Context,
	funcs ...AsyncFn,
) ([][]any, error) {
	return allCompleted(ctx, funcs...)
}

// allCompleted executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout).
func allCompleted(
	parent context.Context,
	funcs ...AsyncFn,
) ([][]any, error) {
	validateAsyncFuncs(funcs...)

	out := make([][]any, len(funcs))
	if len(funcs) == 0 {
		return out, nil
	}

	parent = getContext(parent)

	errs := make([]error, len(funcs))
	errNum := atomic.Int32{}

	wg := sync.WaitGroup{}
	wg.Add(len(funcs))

	for i := 0; i < len(funcs); i++ {
		go func(n int) {
			fn := funcs[n]

			childCtx, childCanFunc := context.WithCancel(parent)
			defer childCanFunc()
			defer wg.Done()

			ret, err := invokeAsyncFn(fn, childCtx, nil)
			if err != nil {
				errNum.Add(1)
				errs[n] = err
			}
			out[n] = ret
		}(i)
	}

	wg.Wait()
	if errNum.Load() == 0 {
		return out, nil
	}

	err := convertErrorListToExecutionErrors(errs, int(errNum.Load()))

	return out, err
}

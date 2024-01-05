package async

import (
	"context"
	"sync"

	"github.com/ghosind/utils"
)

// All executes the functions asynchronously until all functions have been finished. If some
// function returns an error or panic, it will immediately return the index of the function and the
// error, and send a cancel signal to all other functions by context.
//
// The index of the function will be -1 if all functions have been completed without error or
// panic.
func All(funcs ...AsyncFn) (int, error) {
	return all(context.Background(), funcs...)
}

// AllWithContext executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout). If some function returns an error or panic, it will
// immediately return the index of the index and the error and send a cancel signal to all other
// functions by context.
//
// The index of the function will be -1 if all functions have been completed without error or
// panic, or the context has been canceled (or timeout) before all functions finished.
func AllWithContext(ctx context.Context, funcs ...AsyncFn) (int, error) {
	return all(ctx, funcs...)
}

// all executes the functions asynchronously until all functions have been finished, or the context
// is done (canceled or timeout).
func all(parent context.Context, funcs ...AsyncFn) (int, error) {
	if len(funcs) == 0 {
		return -1, nil
	}

	parent = getContext(parent)

	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	ch := make(chan executeResult, len(funcs))
	defer close(ch)

	for i := 0; i < len(funcs); i++ {
		go runTaskInAll(ctx, i, funcs[i], ch)
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

// runTaskInAll runs the specified function for All / AllWithContext.
func runTaskInAll(ctx context.Context, n int, fn AsyncFn, ch chan<- executeResult) {
	childCtx, childCanFunc := context.WithCancel(ctx)
	defer childCanFunc()

	err := utils.Try(func() error {
		return fn(childCtx)
	})

	select {
	case <-ctx.Done():
		return
	default:
		ch <- executeResult{
			Error: err,
			Index: n,
		}
	}
}

// AllCompleted executes the functions asynchronously until all functions have been finished. It
// will return an error slice that is ordered by the functions order, and a boolean value to
// indicate whether any functions return an error or panic.
func AllCompleted(funcs ...AsyncFn) ([]error, bool) {
	return allCompleted(context.Background(), funcs...)
}

// AllCompletedWithContext executes the functions asynchronously until all functions have been
// finished, or the context is done (canceled or timeout). It will return an error slice that is
// ordered by the functions order, and a boolean value to indicate whether any functions return an
// error or panic.
func AllCompletedWithContext(
	ctx context.Context,
	funcs ...AsyncFn,
) ([]error, bool) {
	return allCompleted(ctx, funcs...)
}

// allCompleted executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout).
func allCompleted(
	parent context.Context,
	funcs ...AsyncFn,
) (errs []error, hasError bool) {
	hasError = false
	errs = make([]error, len(funcs))
	if len(funcs) == 0 {
		return
	}

	parent = getContext(parent)

	wg := sync.WaitGroup{}
	wg.Add(len(funcs))

	for i := 0; i < len(funcs); i++ {
		go func(n int) {
			fn := funcs[n]

			childCtx, childCanFunc := context.WithCancel(parent)
			defer childCanFunc()
			defer wg.Done()

			err := utils.Try(func() error {
				return fn(childCtx)
			})
			if err != nil {
				hasError = true
				errs[n] = err
			}
		}(i)
	}

	wg.Wait()

	return
}

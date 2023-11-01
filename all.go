package async

import (
	"context"
	"errors"
	"sync"
)

// All executes the functions asynchronously until all functions have been finished. If some
// function returns an error or panic, it will return the error immediately and send a cancel
// signal to all other functions by context.
func All(funcs ...func(context.Context) error) error {
	return all(context.Background(), funcs...)
}

// AllWithContext executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout). If some function returns an error or panic, it will
// return the error immediately and send a cancel signal to all other functions by context.
func AllWithContext(ctx context.Context, funcs ...func(context.Context) error) error {
	return all(ctx, funcs...)
}

// all executes the functions asynchronously until all functions have been finished, or the context
// is done (canceled or timeout).
func all(parent context.Context, funcs ...func(context.Context) error) error {
	if len(funcs) == 0 {
		return nil
	}

	if parent == nil {
		parent = context.Background()
	}

	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	errCh := make(chan error)
	defer close(errCh)

	for i := 0; i < len(funcs); i++ {
		fn := funcs[i]
		go func() {
			childCtx, childCanFunc := context.WithCancel(ctx)
			defer childCanFunc()

			err := executionContainer(func() error {
				return fn(childCtx)
			})

			select {
			case <-ctx.Done():
				return
			default:
				errCh <- err
			}
		}()
	}

	finished := 0
	for finished < len(funcs) {
		select {
		case <-parent.Done():
			return errors.New("context canceled")
		case err := <-errCh:
			if err != nil {
				return err
			}
			finished++
		}
	}

	return nil
}

// AllCompleted executes the functions asynchronously until all functions have been finished. It
// will return an error slice that is ordered by the functions order, and a boolean value to
// indicate whether any functions return an error or panic.
func AllCompleted(funcs ...func(context.Context) error) ([]error, bool) {
	return allCompleted(context.Background(), funcs...)
}

// AllCompletedWithContext executes the functions asynchronously until all functions have been
// finished, or the context is done (canceled or timeout). It will return an error slice that is
// ordered by the functions order, and a boolean value to indicate whether any functions return an
// error or panic.
func AllCompletedWithContext(
	ctx context.Context,
	funcs ...func(context.Context) error,
) ([]error, bool) {
	return allCompleted(ctx, funcs...)
}

// allCompleted executes the functions asynchronously until all functions have been finished, or
// the context is done (canceled or timeout).
func allCompleted(
	parent context.Context,
	funcs ...func(context.Context) error,
) (errs []error, hasError bool) {
	hasError = false
	errs = make([]error, len(funcs))
	if len(funcs) == 0 {
		return
	}

	if parent == nil {
		parent = context.Background()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(funcs))

	for i := 0; i < len(funcs); i++ {
		n := i
		fn := funcs[n]
		go func() {
			childCtx, childCanFunc := context.WithCancel(parent)
			defer childCanFunc()
			defer wg.Done()

			err := executionContainer(func() error {
				return fn(childCtx)
			})
			if err != nil {
				hasError = true
				errs[n] = err
			}
		}()
	}

	wg.Wait()

	return
}

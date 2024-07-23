package async

import (
	"context"
)

// Series runs the functions by order and returns the results when all functions are completed.
// Each one runs once the previous function has been completed. If any function panics or returns
// an error, no more functions are run and it will return immediately.
//
//	async.Series(func () {
//		// do first thing
//	}, func () {
//		// do second thing
//	}/*, ...*/)
func Series(funcs ...AsyncFn) ([][]any, error) {
	return series(context.Background(), funcs...)
}

// SeriesWithContext runs the functions by order with the specified context and returns the results
// when all functions are completed. Each one runs once the previous function has been completed.
// If any function panics or returns an error, no more functions are run and it will return
// immediately.
func SeriesWithContext(ctx context.Context, funcs ...AsyncFn) ([][]any, error) {
	return series(ctx, funcs...)
}

// series runs the functions by the order.
func series(ctx context.Context, funcs ...AsyncFn) ([][]any, error) {
	validateAsyncFuncs(funcs...)

	ctx = getContext(ctx)
	ret := make([][]any, len(funcs))

	for i := 0; i < len(funcs); i++ {
		fn := funcs[i]
		out, err := invokeAsyncFn(fn, ctx, nil)
		ret[i] = out
		if err != nil {
			return ret, &executionError{
				index: i,
				err:   err,
			}
		}
	}

	return ret, nil
}

package async

import (
	"context"
	"sync/atomic"
)

// Race executes the functions asynchronously, it will return the index and the result of the first
// of the finished function (including panic), and it will not send a cancel signal to other
// functions.
//
//	out, index, err := async.Race(func(ctx context.Context) (int, error) {
//	  request.Get("https://example.com")
//	  return 0, nil
//	}, func(ctx context.Context) (string, error) {
//	  time.Sleep(500 * time.Millisecond)
//	  return "test", nil
//	})
//	// If the first function faster than the second one:
//	// out: []any{0, <nil>}, index: 0, err: <nil>
//	//
//	// Otherwise:
//	// out: []any{"test", <nil>}, index: 1, err: <nil>
func Race(funcs ...AsyncFn) ([]any, int, error) {
	return race(context.Background(), funcs...)
}

// RaceWithContext executes the functions asynchronously, it will return the index and the result
// of the first of the finished function (including panic), and it will not send a cancel signal
// to other functions.
func RaceWithContext(ctx context.Context, funcs ...AsyncFn) ([]any, int, error) {
	return race(ctx, funcs...)
}

// race executes the functions asynchronously, it will return the index and the result of the first
// of the finished function (including panic).
func race(ctx context.Context, funcs ...AsyncFn) ([]any, int, error) {
	if len(funcs) == 0 {
		return nil, -1, nil
	}
	validateAsyncFuncs(funcs...)

	ctx = getContext(ctx)

	finished := atomic.Bool{}
	ch := make(chan executeResult)
	defer close(ch)

	for i := 0; i < len(funcs); i++ {
		go func(n int) {
			fn := funcs[n]

			ret, err := invokeAsyncFn(fn, ctx, nil)
			if finished.CompareAndSwap(false, true) {
				ch <- executeResult{
					Index: n,
					Error: err,
					Out:   ret,
				}
			}
		}(i)
	}

	ret := <-ch
	if ret.Error != nil {
		return ret.Out, ret.Index, &executionError{
			index: ret.Index,
			err:   ret.Error,
		}
	}

	return ret.Out, ret.Index, nil
}

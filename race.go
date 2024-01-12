package async

import (
	"context"
	"sync/atomic"
)

// Race executes the functions asynchronously, it will return the index and the result of the first
// of the finished function (including panic), and it will not send a cancel signal to other
// functions.
func Race(funcs ...AsyncFn) (int, error) {
	return race(context.Background(), funcs...)
}

// RaceWithContext executes the functions asynchronously, it will return the index and the result
// of the first of the finished function (including panic), and it will not send a cancel signal
// to other functions.
func RaceWithContext(ctx context.Context, funcs ...AsyncFn) (int, error) {
	return race(ctx, funcs...)
}

// race executes the functions asynchronously, it will return the index and the result of the first
// of the finished function (including panic).
func race(ctx context.Context, funcs ...AsyncFn) (int, error) {
	if len(funcs) == 0 {
		return -1, nil
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

	return ret.Index, ret.Error
}

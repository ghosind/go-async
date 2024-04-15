package async

import "context"

// Times executes the function n times, and returns the results. It'll terminate if any function
// panics or returns an error.
//
//	// Calls api 5 times.
//	async.Times(5, func () => error {
//		return CallAPI()
//	})
func Times(n int, fn AsyncFn) ([][]any, error) {
	return times(context.Background(), n, 0, fn)
}

// TimesWithContext executes the function n times with the context, and returns the results. It'll
// terminate if any function panics or returns an error.
func TimesWithContext(ctx context.Context, n int, fn AsyncFn) ([][]any, error) {
	return times(ctx, n, 0, fn)
}

// TimesLimit executes the function n times with the specified concurrency limit, and returns the
// results. It'll terminate if any function panics or returns an error.
//
//	// Calls api 5 times with 2 concurrency.
//	async.TimesLimit(5, 2, func () {
//		return CallAPI()
//	})
func TimesLimit(n, concurrency int, fn AsyncFn) ([][]any, error) {
	return times(context.Background(), n, concurrency, fn)
}

// TimesLimitWithContext executes the function n times with the specified concurrency limit and
// the context, and returns the results. It'll terminate if any function panics or returns an
// error.
func TimesLimitWithContext(ctx context.Context, n, concurrency int, fn AsyncFn) ([][]any, error) {
	return times(ctx, n, concurrency, fn)
}

// TimesSeries executes the function n times with only a single invocation at a time, and returns
// the results. It'll terminate if any function panics or returns an error.
//
//	// Calls api 5 times but runs one at a time.
//	async.TimesSeries(5, func () {
//		return CallAPI()
//	})
func TimesSeries(n int, fn AsyncFn) ([][]any, error) {
	return times(context.Background(), n, 1, fn)
}

// TimesSeriesWithContext executes the function n times with the context and only a single
// invocation at a time, and returns the results. It'll terminate if any function panics or
// returns an error.
func TimesSeriesWithContext(ctx context.Context, n int, fn AsyncFn) ([][]any, error) {
	return times(ctx, n, 1, fn)
}

// times executes the function n times withe the specified concurrency.
func times(parent context.Context, n, concurrency int, fn AsyncFn) ([][]any, error) {
	// the number of concurrency should be 0 (no limitation) or greater than 0.
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}
	validateAsyncFuncs(fn)

	parent = getContext(parent)
	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	out := make([][]any, n)

	ch := make(chan executeResult, n)
	var conch chan empty

	if concurrency > 0 {
		conch = make(chan empty, concurrency)
	}

	go func() {
		for i := 0; i < n; i++ {
			if conch != nil {
				conch <- empty{}
			}

			go runTaskInParallel(ctx, i, fn, conch, ch)
		}
	}()

	finished := 0
	for finished < n {
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

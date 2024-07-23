package async

import (
	"context"
	"sync"
)

// ForeverFn is the function to run in the Forever function.
type ForeverFn func(ctx context.Context, next func(context.Context)) error

// Forever runs the function indefinitely until the function panics or returns an error.
//
// You can use the context and call the next function to pass values to the next invocation. The
// next function can be invoked one time only, and it will have no effect if it is invoked again.
//
//	err := async.Forever(func(ctx context.Context, next func(context.Context)) error {
//	  v := ctx.Value("key")
//	  if v != nil {
//	    vi := v.(int)
//	    if vi == 3 {
//	      return errors.New("finish")
//	    }
//
//	    fmt.Printf("value: %d\n", vi)
//
//	    next(context.WithValue(ctx, "key", vi+1))
//	  } else {
//	    next(context.WithValue(ctx, "key", 1))
//	  }
//
//	  return nil
//	})
//	fmt.Printf("err: %v\n", err)
//	// value: 1
//	// value: 2
//	// err: finish
func Forever(fn ForeverFn) error {
	return forever(context.Background(), fn)
}

// ForeverWithContext runs the function indefinitely until the function panics or returns an error.
//
// You can use the context and call the next function to pass values to the next invocation. The
// next function can be invoked one time only, and it will have no effect if it is invoked again.
func ForeverWithContext(ctx context.Context, fn ForeverFn) error {
	return forever(ctx, fn)
}

// forever runs the function indefinitely.
func forever(
	parent context.Context,
	fn ForeverFn,
) error {
	validateAsyncFuncs(fn)
	ctx := getContext(parent)
	nextCtx := ctx

	for {
		once := sync.Once{}
		next := func(ctx context.Context) {
			once.Do(func() {
				nextCtx = ctx
			})
		}

		ctx = nextCtx

		_, err := invokeAsyncFn(fn, ctx, []any{next})
		if err != nil {
			return err
		}
	}
}

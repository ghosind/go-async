package async

import (
	"context"
	"sync"

	"github.com/ghosind/utils"
)

// ForeverFn is the function to run in the Forever function.
type ForeverFn func(ctx context.Context, next func(context.Context)) error

// Forever runs the function indefinitely until the function panics or returns an error.
//
// You can use the context and call the next function to pass values to the next invocation. The
// next function can be invoked one time only, and it will have no effect if it is invoked again.
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

		err := utils.Try(func() error {
			return fn(ctx, next)
		})
		if err != nil {
			return err
		}
	}
}

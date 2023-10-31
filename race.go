package async

import (
	"context"
	"sync/atomic"
)

// Race executes the functions asynchronously, it will return the result of the first of the
// finished function (including panic), and it will not send a cancel signal to other functions.
func Race(funcs ...func(context.Context) error) error {
	return race(context.Background(), funcs...)
}

// RaceWithContext executes the functions asynchronously, it will return the result of the first of
// the finished function (including panic), and it will not send a cancel signal to other
// functions.
func RaceWithContext(ctx context.Context, funcs ...func(context.Context) error) error {
	return race(ctx, funcs...)
}

// race executes the functions asynchronously, it will return the result of the first of the
// finished function (including panic).
func race(ctx context.Context, funcs ...func(context.Context) error) error {
	if len(funcs) == 0 {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	finished := atomic.Bool{}
	ch := make(chan error)
	defer close(ch)

	for i := 0; i < len(funcs); i++ {
		fn := funcs[i]

		go func() {
			err := fn(ctx)
			if finished.CompareAndSwap(false, true) {
				ch <- err
			}
		}()
	}

	err := <-ch

	return err
}

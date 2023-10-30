package async

import (
	"context"
	"sync/atomic"
)

func Race(funcs ...func(context.Context) error) error {
	return race(context.Background(), funcs...)
}

func RaceWithContext(ctx context.Context, funcs ...func(context.Context) error) error {
	return race(ctx, funcs...)
}

func race(ctx context.Context, funcs ...func(context.Context) error) error {
	if len(funcs) == 0 {
		return nil
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

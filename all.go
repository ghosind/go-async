package async

import (
	"context"
	"errors"
)

func All(funcs ...func(context.Context) error) error {
	return all(context.Background(), funcs...)
}

func AllWithContext(ctx context.Context, funcs ...func(context.Context) error) error {
	return all(ctx, funcs...)
}

func all(parent context.Context, funcs ...func(context.Context) error) error {
	if len(funcs) == 0 {
		return nil
	}

	ctx, canFunc := context.WithCancel(parent)
	errCh := make(chan error)
	retCh := make(chan struct{}, len(funcs))

	defer canFunc()
	defer close(errCh)
	defer close(retCh)

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
				if err != nil {
					errCh <- err
				} else {
					retCh <- struct{}{}
				}
			}
		}()
	}

	finished := 0
	for {
		select {
		case <-parent.Done():
			return errors.New("context canceled")
		case err := <-errCh:
			return err
		case <-retCh:
			finished++
		}

		if finished == len(funcs) {
			return nil
		}
	}
}

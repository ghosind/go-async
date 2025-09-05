package async

import "context"

// Fallback tries to run the functions in order until one function does not panic or return an
// error. It returns nil if one function succeeds, or returns the last error if all functions fail.
//
//	err := async.Fallback(func() error {
//	  return errors.New("first error")
//	}, func() error {
//	  return errors.New("second error")
//	}, func() error {
//	  return nil
//	}, func() error {
//	  return errors.New("third error")
//	})
//	// err: <nil>
func Fallback(fn AsyncFn, fallbacks ...AsyncFn) error {
	return fallback(context.Background(), fn, fallbacks...)
}

// FallbackWithContext tries to run the functions in order with the specified context until one
// function does not panic or return an error. It returns nil if one function succeeds, or returns
// the last error if all functions fail.
func FallbackWithContext(ctx context.Context, fn AsyncFn, fallbacks ...AsyncFn) error {
	return fallback(ctx, fn, fallbacks...)
}

// fallback runs the functions in order until one function does not panic or return an error.
func fallback(parent context.Context, fn AsyncFn, fallbacks ...AsyncFn) error {
	funcs := append([]AsyncFn{fn}, fallbacks...)
	ctx := getContext(parent)
	var err error

	for _, fn := range funcs {
		_, err = invokeAsyncFn(fn, ctx, nil)
		if err == nil {
			return nil
		}
	}

	return err
}

package async

import "context"

// Seq runs the functions in order, and each function consumes the returns value of the previous
// function. It returns the result of the last function, or it terminates and returns the error
// that panics or returns by the function in the list.
//
//	out, err :=Seq(func () int {
//		return 1
//	}, func (n int) int {
//		return n + 1
//	})
//	// out: [2]
//	// err: <nil>
func Seq(funcs ...AsyncFn) ([]any, error) {
	return seq(context.Background(), funcs...)
}

// Seq runs the functions in order with the specified context, and each function consumes the
// returns value of the previous function. It returns the result of the last function, or it
// terminates and returns the error that panics or returns by the function in the list.
func SeqWithContext(ctx context.Context, funcs ...AsyncFn) ([]any, error) {
	return seq(ctx, funcs...)
}

// seq runs the functions in order, and each function consumes the return values of the previous
// function.
func seq(ctx context.Context, funcs ...AsyncFn) ([]any, error) {
	// TODO: validate functions

	var ret []any
	ctx = getContext(ctx)

	for i, fn := range funcs {
		out, err := invokeAsyncFn(fn, ctx, ret)
		if err != nil {
			return nil, &executionError{
				index: i,
				err:   err,
			}
		}
		ret = out
	}

	return ret, nil
}

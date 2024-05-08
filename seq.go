package async

import (
	"context"
	"reflect"
)

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
	if err := validateSeqFuncs(funcs...); err != nil {
		return nil, err
	}

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

// validateSeqFuncs checks the functions for in seq functions list.
func validateSeqFuncs(funcs ...AsyncFn) error {
	types := make([]reflect.Type, 0, len(funcs))
	for _, fn := range funcs {
		if fn == nil {
			return ErrNotFunction
		}
		ty := reflect.TypeOf(fn)
		if ty.Kind() != reflect.Func {
			return ErrNotFunction
		}
		types = append(types, ty)
	}

	for i := 1; i < len(types); i++ {
		err := validateSeqFuncParams(types[i-1], types[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// validateSeqFuncParams checks the previous function's return values and the current function's
// parameters, and returns an error if they are not match.
func validateSeqFuncParams(prev, cur reflect.Type) error {
	isTakeContext := isFuncTakesContext(cur)
	numIn := cur.NumIn()
	numOut := prev.NumOut()

	if isTakeContext {
		numIn--
	}
	if prev.NumOut() < numIn {
		return ErrInvalidSeqFuncs
	}

	i := 0
	j := 0

	if isTakeContext {
		if numOut > 0 && isContextType(prev.Out(0)) {
			i++
		}
		numIn++
		j++
	}
	for i < numOut && j < numIn {
		if prev.Out(i) != cur.In(j) {
			return ErrInvalidSeqFuncs
		}
		i++
		j++
	}

	return nil
}

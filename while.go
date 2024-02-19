package async

import (
	"context"
	"reflect"
)

// While repeatedly calls the function while the test function returns true. A valid test function
// must match the following requirements.
//
// - The first return value of the test function must be a boolean value.
// - It should have no parameters or accept a context only.
//
//	c := 0
//	While(func() bool {
//	  return c == 5
//	}, func() {
//	  c++
//	})
func While(testFn, fn AsyncFn) ([]any, error) {
	return while(context.Background(), testFn, fn)
}

// WhileWithContext repeatedly calls the function with the specified context while the test
// function returns true.
func WhileWithContext(ctx context.Context, testFn, fn AsyncFn) ([]any, error) {
	return while(ctx, testFn, fn)
}

// while repeatedly calls the function while the test function returns true.
func while(parent context.Context, testFn, fn AsyncFn) ([]any, error) {
	validateWhileFuncs(testFn, fn)

	ctx := getContext(parent)
	var out []any
	var err error

	for {
		testOut, testErr := invokeAsyncFn(testFn, ctx, nil)
		if testErr != nil {
			return out, testErr
		}

		isContinue := testOut[0].(bool)
		if !isContinue {
			return out, nil
		}

		out, err = invokeAsyncFn(fn, ctx, nil)
		if err != nil {
			break
		}
	}

	return out, err
}

// validateWhileFuncs validates the test function and the execution function for while.
func validateWhileFuncs(testFn, fn AsyncFn) {
	if testFn == nil || fn == nil {
		panic(ErrNotFunction)
	}
	tft := reflect.TypeOf(testFn) // reflect.Type of the test function
	if tft.Kind() != reflect.Func || reflect.TypeOf(fn).Kind() != reflect.Func {
		panic(ErrNotFunction)
	}

	if tft.NumOut() <= 0 || tft.Out(0).Kind() != reflect.Bool {
		panic(ErrInvalidTestFunc)
	}

	numIn := tft.NumIn()
	isTakeContext := isFuncTakesContext(tft)
	if isTakeContext {
		numIn--
	}
	if numIn != 0 {
		panic(ErrInvalidTestFunc)
	}
}

package async

import (
	"context"
	"reflect"
)

// Until repeatedly calls the function until the test function returns true. A valid test function
// must match the following requirements.
//
// - The first return value of the test function must be a boolean value.
// - The parameters' number of the test function must be equal to the return values' number of the
// execution function (exclude context).
// - The parameters' types of the test function must be the same or convertible to the return
// values' types of the execution function.
//
//	c := 0
//	Until(func() bool {
//	  return c == 5
//	}, func() {
//	  c++
//	})
func Until(testFn, fn AsyncFn) ([]any, error) {
	return until(context.Background(), testFn, fn)
}

// UntilWithContext repeatedly calls the function with the specified context until the test
// function returns true.
func UntilWithContext(ctx context.Context, testFn, fn AsyncFn) ([]any, error) {
	return until(ctx, testFn, fn)
}

// until repeatedly calls the function until the test function returns true.
func until(parent context.Context, testFn, fn AsyncFn) ([]any, error) {
	isNoParam := validateUntilFuncs(testFn, fn)

	ctx := getContext(parent)

	for {
		out, _ := invokeAsyncFn(fn, ctx, nil)

		params := out
		if isNoParam {
			params = nil
		}
		testOut, testErr := invokeAsyncFn(testFn, ctx, params)
		if testErr != nil {
			return out, testErr
		}

		isContinue := testOut[0].(bool)
		if !isContinue {
			return out, nil
		}
	}
}

// validateUntilFuncs validates the test function and the execution function.
func validateUntilFuncs(testFn, fn AsyncFn) (isNoParam bool) {
	if testFn == nil || fn == nil {
		panic(ErrNotFunction)
	}
	tft := reflect.TypeOf(testFn) // reflect.Type of the test function
	ft := reflect.TypeOf(fn)      // reflect.Type of the function
	if tft.Kind() != reflect.Func || ft.Kind() != reflect.Func {
		panic(ErrNotFunction)
	}

	if tft.NumOut() <= 0 || tft.Out(0).Kind() != reflect.Bool {
		panic(ErrInvalidTestFunc)
	}

	ii := 0 // index of the test function input parameters list
	oi := 0 // index of the function return values list
	numIn := tft.NumIn()
	isTakeContext := isFuncTakesContext(tft)
	if isTakeContext {
		numIn--
		ii++
	}
	if numIn != 0 && numIn != ft.NumOut() {
		panic(ErrInvalidTestFunc)
	}
	if numIn == 0 {
		return true
	}

	for oi < numIn {
		it := tft.In(ii) // type of the value in the test function input parameters list
		ot := ft.Out(oi) // type of the value in the function return values list

		if it != ot && !it.ConvertibleTo(ot) {
			panic(ErrInvalidTestFunc)
		}

		ii++
		oi++
	}

	return false
}

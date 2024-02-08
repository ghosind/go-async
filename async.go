package async

import (
	"context"
	"reflect"

	"github.com/ghosind/utils"
)

// AsyncFn is the function to run, the function can be a function without any restriction that accepts any parameters and any return values. For the best practice, please define the function like the following styles:
//
//	func(context.Context) error
//	func(context.Context) (out_type, error)
//	func(context.Context, in_type) error
//	func(context.Context, in_type) (out_type, error)
//	func(context.Context, in_type1, in_type2/*, ...*/) (out_type1, out_type_2,/* ...,*/ error)
type AsyncFn any

// executeResult indicates the execution result whether the function returns an error or panic, and
// the index of the function in the parameters list.
type executeResult struct {
	// Error is the execution result of the function, it will be nil if the function does not return
	// an error and does not panic.
	Error error
	// Index is the index of the function in the parameters list.
	Index int
	// Out is an array to store the return values without the last error.
	Out []any
}

// empty is a smallest cost struct.
type empty struct{}

// contextType is the reflect type of context.Context.
var contextType reflect.Type = reflect.TypeOf((*context.Context)(nil)).Elem()

// errorType is the reflect type of an error
var errorType reflect.Type = reflect.TypeOf((*error)(nil)).Elem()

// getContext returns the specified non-nil context from the parameter, or creates and returns a
// new empty context.
func getContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}

	return context.Background()
}

// validateAsyncFuncs validates the functions list, and it will panic if any function is nil or not
// a function.
func validateAsyncFuncs(funcs ...AsyncFn) {
	for _, fn := range funcs {
		if fn == nil || reflect.TypeOf(fn).Kind() != reflect.Func {
			panic(ErrNotFunction)
		}
	}
}

// isFuncTakesContext checks the function takes a Context as the first argument.
func isFuncTakesContext(fn reflect.Type) bool {
	if fn.NumIn() <= 0 {
		return false
	}

	in := fn.In(0)

	if in.Kind() != reflect.Interface || !in.Implements(contextType) || !contextType.Implements(in) {
		return false
	}

	return true
}

// isFuncReturnsError checks the last return value of the function is an error if the function
// returns some values.
func isFuncReturnsError(fn reflect.Type) bool {
	if fn.NumOut() <= 0 {
		return false
	}

	out := fn.Out(fn.NumOut() - 1)

	if out.Kind() != reflect.Interface || !out.Implements(errorType) || !errorType.Implements(out) {
		return false
	}

	return true
}

// invokeAsyncFn tries to call the function with the specified parameters, and it'll also set the
// context if it is the function's first parameter. After the function is finished, it will return
// a return values array and the error. It will store the return values into the out array without
// the error if it is the last return value.
func invokeAsyncFn(fn AsyncFn, ctx context.Context, params []any) ([]any, error) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	var out []reflect.Value

	in := makeFuncIn(ft, ctx, params)

	numRet := ft.NumOut()
	ret := make([]any, 0, numRet)

	err := utils.Try(func() error {
		out = fv.Call(in)
		return nil
	})
	if err != nil {
		return ret, err
	}

	if isFuncReturnsError(ft) {
		if out[numRet-1].IsNil() {
			err = nil
		} else {
			err = out[numRet-1].Interface().(error)
		}
	}
	for i := 0; i < numRet; i++ {
		ret = append(ret, out[i].Interface())
	}

	return ret, err
}

// makeFuncIn checks the parameters of the function and the params slice from the caller, and
// returns a reflect.Value slice of the input parameters. It'll prepend the context to the
// parameter list if the function's first parameter is a context.
//
// The function will panic an unmatched param error if the number of parameters is not equal to the
// param list, or some elements' types of parameters are not match.
func makeFuncIn(ft reflect.Type, ctx context.Context, params []any) []reflect.Value {
	numIn := ft.NumIn()
	isTakeContext := isFuncTakesContext(ft)
	if isTakeContext {
		numIn--
	}
	if numIn != len(params) {
		panic(ErrUnmatchedParam)
	}

	in := make([]reflect.Value, ft.NumIn())
	i := 0 // index of the input parameter list

	if isTakeContext {
		// prepend context to the input parameter list
		in[i] = reflect.ValueOf(ctx)
		i++
	}

	for _, v := range params {
		vv := reflect.ValueOf(v)
		vt := vv.Type() // the type of the value
		it := ft.In(i)  // the type in the parameter list

		if vt != it {
			// if the value's type does not match the parameter list, try to convert it first
			if vt.ConvertibleTo(it) {
				vv = vv.Convert(it)
			} else {
				panic(ErrUnmatchedParam)
			}
		}

		in[i] = vv
		i++
	}

	return in
}

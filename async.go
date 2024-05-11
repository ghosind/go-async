package async

import (
	"context"
	"reflect"

	"github.com/ghosind/go-try"
)

// AsyncFn is the function to run, the function can be a function without any restriction that
// accepts any parameters and any return values. For the best practice, please define the function
// like the following styles:
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

// isContextType returns a boolean value to indicates whether the type is context or not.
func isContextType(ty reflect.Type) bool {
	return ty.Kind() == reflect.Interface &&
		ty.Implements(contextType) && contextType.Implements(ty)
}

// isFuncTakesContexts checks the function takes Contexts as the arguments.
func isFuncTakesContexts(fn reflect.Type) (bool, int) {
	if fn.NumIn() <= 0 {
		return false, 0
	}

	hasContext := false
	contextNum := 0
	for i := 0; i < fn.NumIn(); i++ {
		ok := isContextType(fn.In(i))
		if ok {
			hasContext = true
			contextNum++
		} else {
			break
		}
	}

	return hasContext, contextNum
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

// isFirstParamContext checks the any type slice, and return true if the first element in the slice
// is a context object.
func isFirstParamContext(params []any, numIn int) bool {
	if len(params) == 0 || len(params) < numIn {
		return false
	}

	ty := reflect.TypeOf(params[0])
	return ty == nil || ty.Implements(contextType)
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
	ret := make([]any, numRet)

	_, err := try.Try(func() {
		out = fv.Call(in)
	})
	if err != nil {
		for i := 0; i < numRet; i++ {
			ret[i] = reflect.Zero(ft.Out(i)).Interface()
		}
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
		ret[i] = out[i].Interface()
	}

	return ret, err
}

// makeFuncIn makes a reflected values list of the parameters to call the function.
func makeFuncIn(ft reflect.Type, ctx context.Context, params []any) []reflect.Value {
	isTakeContext, _ := isFuncTakesContexts(ft)
	isContextParam := isTakeContext && isFirstParamContext(params, ft.NumIn())

	if !ft.IsVariadic() {
		return makeNonVariadicFuncIn(ft, ctx, params, isTakeContext, isContextParam)
	} else {
		panic("variadic function unsupported")
	}
}

// makeNonVariadicFuncIn checks the parameters of the non-variadic function and the params slice
// from the caller, and returns a reflect.Value slice of the input parameters. It'll prepend the
// context to the parameter list if the function's first parameter is a context and the first
// element in the parameter list is not a context.
//
// The function will panic an unmatched param error if the number of parameters for the function is
// greater to the specified parameters list, or some elements' types of parameters are not match.
func makeNonVariadicFuncIn(
	ft reflect.Type,
	ctx context.Context,
	params []any,
	isTakeContext, isContextParam bool,
) []reflect.Value {
	numIn := ft.NumIn()
	if isTakeContext && !isContextParam {
		numIn--
	}
	if numIn > len(params) {
		panic(ErrUnmatchedParam)
	}

	in := make([]reflect.Value, ft.NumIn())
	i := 0 // index of the input parameter list

	if isTakeContext && !isContextParam {
		// prepend context to the input parameter list
		in[i] = reflect.ValueOf(ctx)
		i++
		numIn++
	}

	for j := 0; i < numIn; j++ {
		v := params[j]
		vt := reflect.TypeOf(v) // the type of the value
		vv := reflect.ValueOf(v)
		it := ft.In(i) // the type in the parameter list

		if vt != it {
			// if the value's type does not match the parameter list, try to convert it first
			if vt != nil && vt.ConvertibleTo(it) {
				vv = vv.Convert(it)
			} else if v == nil {
				// check the parameter's type is whether nil-able or not when the value is nil
				kind := it.Kind()
				switch kind {
				case reflect.Chan, reflect.Map, reflect.Pointer, reflect.UnsafePointer,
					reflect.Interface, reflect.Slice:
					vv = reflect.Zero(it)
				default:
					panic(ErrUnmatchedParam)
				}
			} else {
				panic(ErrUnmatchedParam)
			}
		}

		in[i] = vv
		i++
	}

	return in
}

// isValidNextFunc checks the current function's return values and the next function's parameters,
// and returns a boolean value to indicates whether the functions are match or not
func isValidNextFunc(cur, next reflect.Type) bool {
	isTakeContext, _ := isFuncTakesContexts(next)
	numOut := cur.NumOut()
	numIn := next.NumIn()

	if isTakeContext {
		numIn--
	}
	if numOut < numIn {
		return false
	}

	i := 0
	j := 0

	if isTakeContext {
		if numOut > 0 && isContextType(cur.Out(0)) {
			i++
		}
		numIn++
		j++
	}
	for i < numOut && j < numIn {
		if cur.Out(i) != next.In(j) {
			return false
		}
		i++
		j++
	}

	return true
}

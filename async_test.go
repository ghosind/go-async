package async

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ghosind/go-assert"
)

func TestGetContext(t *testing.T) {
	a := assert.New(t)

	todoCtx := context.TODO()
	ctx := getContext(todoCtx)
	a.EqualNow(ctx, todoCtx)

	//lint:ignore SA1012 for test case only
	ctx = getContext(nil)
	a.NotNilNow(ctx)
	a.NotEqualNow(ctx, todoCtx)
}

func TestValidateAsyncFuncs(t *testing.T) {
	a := assert.New(t)

	a.NotPanicNow(func() {
		validateAsyncFuncs()
	})
	a.NotPanicNow(func() {
		validateAsyncFuncs(func() {})
	})
	a.PanicOfNow(func() {
		validateAsyncFuncs(func() {}, nil, func() {})
	}, ErrNotFunction)
	a.PanicOfNow(func() {
		validateAsyncFuncs(func() {}, 1, func() {})
	}, ErrNotFunction)
}

func TestIsFuncTakesContexts(t *testing.T) {
	a := assert.New(t)

	isTakeContext, contextNum := isFuncTakesContexts(reflect.TypeOf(func(context.Context) {}))
	a.TrueNow(isTakeContext)
	a.EqualNow(contextNum, 1)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func(context.Context, int) {}))
	a.TrueNow(isTakeContext)
	a.EqualNow(contextNum, 1)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func(context.Context, context.Context, int) {}))
	a.TrueNow(isTakeContext)
	a.EqualNow(contextNum, 2)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func(context.Context, int, context.Context) {}))
	a.TrueNow(isTakeContext)
	a.EqualNow(contextNum, 1)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func() {}))
	a.NotTrueNow(isTakeContext)
	a.EqualNow(contextNum, 0)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func(int) {}))
	a.NotTrueNow(isTakeContext)
	a.EqualNow(contextNum, 0)

	isTakeContext, contextNum = isFuncTakesContexts(reflect.TypeOf(func(int, context.Context) {}))
	a.NotTrueNow(isTakeContext)
	a.EqualNow(contextNum, 0)
}

func TestIsFuncReturnsError(t *testing.T) {
	a := assert.New(t)

	a.TrueNow(isFuncReturnsError(reflect.TypeOf(func() error { return nil })))
	a.TrueNow(isFuncReturnsError(reflect.TypeOf(func() (int, error) { return 0, nil })))
	a.TrueNow(isFuncReturnsError(reflect.TypeOf(func() (int, string, error) { return 0, "", nil })))
	a.NotTrueNow(isFuncReturnsError(reflect.TypeOf(func() {})))
	a.NotTrueNow(isFuncReturnsError(reflect.TypeOf(func() int { return 0 })))
	a.NotTrueNow(isFuncReturnsError(reflect.TypeOf(func() (int, string) { return 0, "" })))
	//lint:ignore ST1008 for test
	a.NotTrueNow(isFuncReturnsError(reflect.TypeOf(func() (error, int) { return nil, 0 })))
}

func TestInvokeAsyncFn(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	expectErr := errors.New("expected error")

	ret, err := invokeAsyncFn(func() {}, ctx, nil)
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func() { panic(expectErr) }, ctx, nil)
	a.EqualNow(err, expectErr)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func() error { return nil }, ctx, nil)
	a.NilNow(err)
	a.EqualNow(ret, []any{nil})

	ret, err = invokeAsyncFn(func() error { return expectErr }, ctx, nil)
	a.EqualNow(err, expectErr)
	a.EqualNow(ret, []any{expectErr})

	ret, err = invokeAsyncFn(func() int { return 1 }, ctx, nil)
	a.NilNow(err)
	a.EqualNow(ret, []any{1})

	ret, err = invokeAsyncFn(func() (int, error) { return 1, nil }, ctx, nil)
	a.NilNow(err)
	a.EqualNow(ret, []any{1, nil})

	ret, err = invokeAsyncFn(func() (int, error) { return 1, expectErr }, ctx, nil)
	a.EqualNow(err, expectErr)
	a.EqualNow(ret, []any{1, expectErr})

	ret, err = invokeAsyncFn(func() (int, string, error) { return 1, "test", nil }, ctx, nil)
	a.NilNow(err)
	a.EqualNow(ret, []any{1, "test", nil})

	ret, err = invokeAsyncFn(func(ctx context.Context) {}, ctx, []any{})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(ctx context.Context) {}, ctx, []any{ctx})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(ctx context.Context) {}, ctx, []any{nil})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	// a.PanicOfNow(func() {
	// 	invokeAsyncFn(func(ctx context.Context, vals ...int) {}, ctx, []any{nil})
	// }, "variadic function unsupported")

	ret, err = invokeAsyncFn(func() int {
		panic(expectErr)
	}, ctx, nil)
	a.EqualNow(err, expectErr)
	a.EqualNow(ret, []any{0})
}

func TestInvokeVariadicAsyncFn(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()

	ret, err := invokeAsyncFn(func(vals ...int) {}, ctx, []any{})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(ctx context.Context, vals ...int) {}, ctx, []any{})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(vals ...int) {}, ctx, []any{1, 2, 3})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(ctx context.Context, vals ...int) {}, ctx, []any{ctx, 1, 2, 3})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(ctx context.Context, s string, vals ...int) {}, ctx, []any{"test", 1})
	a.NilNow(err)
	a.EqualNow(ret, []any{})

	ret, err = invokeAsyncFn(func(vals ...int) int {
		if len(vals) > 0 {
			return vals[0]
		}
		return -1
	}, ctx, []any{1, 2, 3})
	a.NilNow(err)
	a.EqualNow(ret, []any{1})

	a.PanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, s string, vals ...int) {}, ctx, []any{})
	})

	a.PanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, s string, vals ...int) {}, ctx, []any{"a", "b"})
	})

	a.PanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, vals ...int) {}, ctx, []any{"a", "b"})
	})
}

func TestInvokeAsyncFnWithParams(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()

	a.NotPanicNow(func() { invokeAsyncFn(func(ctx context.Context) {}, ctx, nil) })
	a.NotPanicNow(func() { invokeAsyncFn(func(n int) {}, ctx, []any{float64(1.0)}) })
	a.NotPanicNow(func() { invokeAsyncFn(func(err error) {}, ctx, []any{nil}) })
	a.NotPanicNow(func() { invokeAsyncFn(func(s *struct{ V int }) {}, ctx, []any{nil}) })
	a.NotPanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{float64(1.0)})
	})
	a.NotPanicNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, []any{float64(1.0), "hello"})
	})
	a.NotPanicNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, []any{float64(1.0), 1})
	})
	a.NotPanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{float64(1.0), "hello"})
	})
	a.NotPanicNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{float64(1.0), 1})
	})

	a.PanicOfNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, nil)
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, []any{nil})
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, []any{"hello"})
	}, ErrUnmatchedParam)

	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, nil)
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{"hello"})
	}, ErrUnmatchedParam)

}

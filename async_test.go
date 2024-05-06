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

func TestIsFuncTakesContext(t *testing.T) {
	a := assert.New(t)

	a.TrueNow(isFuncTakesContext(reflect.TypeOf(func(context.Context) {})))
	a.TrueNow(isFuncTakesContext(reflect.TypeOf(func(context.Context, int) {})))
	a.NotTrueNow(isFuncTakesContext(reflect.TypeOf(func() {})))
	a.NotTrueNow(isFuncTakesContext(reflect.TypeOf(func(int) {})))
	a.NotTrueNow(isFuncTakesContext(reflect.TypeOf(func(int, context.Context) {})))
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

	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, vals ...int) {}, ctx, []any{nil})
	}, "variadic function unsupported")

	ret, err = invokeAsyncFn(func() int {
		panic(expectErr)
	}, ctx, nil)
	a.EqualNow(err, expectErr)
	a.EqualNow(ret, []any{0})
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
		invokeAsyncFn(func(n int) {}, ctx, []any{float64(1.0), "hello"})
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(n int) {}, ctx, []any{float64(1.0), 1})
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, nil)
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{"hello"})
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{float64(1.0), "hello"})
	}, ErrUnmatchedParam)
	a.PanicOfNow(func() {
		invokeAsyncFn(func(ctx context.Context, n int) {}, ctx, []any{float64(1.0), 1})
	}, ErrUnmatchedParam)
}

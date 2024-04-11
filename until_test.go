package async_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestUntil(t *testing.T) {
	a := assert.New(t)
	count := 0

	out, err := async.Until(func(c int) bool {
		return c < 5
	}, func() int {
		count++
		return count
	})
	a.NilNow(err)
	a.EqualNow(out, []any{5})
}

func TestUntilInvalidParameters(t *testing.T) {
	a := assert.New(t)

	a.PanicOfNow(func() {
		async.Until(nil, func() {})
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.Until(func() {}, nil)
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.Until(1, "hello")
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.Until(func() {}, func() {})
	}, async.ErrInvalidTestFunc)
	a.NotPanicNow(func() {
		async.Until(func() bool { return false }, func() {})
	})
	a.NotPanicNow(func() {
		async.Until(func(err error) bool { return false }, func() error { return nil })
	})
	a.NotPanicNow(func() {
		async.Until(func(ctx context.Context, err error) bool { return false }, func() error { return nil })
	})
	a.NotPanicNow(func() {
		async.Until(func(ctx context.Context) bool { return false }, func() error { return nil })
	})
	a.PanicOfNow(func() {
		async.Until(func(ctx context.Context, i int) bool { return false }, func() error { return nil })
	}, async.ErrInvalidTestFunc)
	a.PanicOfNow(func() {
		async.Until(func(ctx context.Context, i int) bool { return false }, func() {})
	}, async.ErrInvalidTestFunc)
}

func TestUntilWithFunctionError(t *testing.T) {
	a := assert.New(t)
	count := 0
	unexpectedErr := errors.New("unexpected error")

	out, err := async.Until(func(c int, err error) bool {
		return c < 5
	}, func() (int, error) {
		count++
		return count, unexpectedErr
	})
	a.NilNow(err)
	a.EqualNow(out, []any{5, unexpectedErr})
}

func TestUntilWithTestFunctionError(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	out, err := async.Until(func(n int) bool {
		panic(expectedErr)
	}, func() int {
		return 0
	})
	a.NotNilNow(err)
	a.EqualNow(err, expectedErr)
	a.EqualNow(out, []any{0})
}

func TestUntilWithContext(t *testing.T) {
	a := assert.New(t)
	ctx, canFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer canFunc()

	start := time.Now()
	out, err := async.UntilWithContext(ctx, func(ctx context.Context) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}, func() {
	})
	a.NilNow(err)
	a.EqualNow(out, []any{})
	dur := time.Since(start)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 150*time.Millisecond)
}

func ExampleUntil() {
	i := 0

	out, err := async.Until(func(n int) bool {
		return n < 3
	}, func() int {
		i++
		return i
	})
	fmt.Println(i)
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// 3
	// [3]
	// <nil>
}

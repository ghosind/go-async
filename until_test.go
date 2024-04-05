package async

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestUntil(t *testing.T) {
	a := assert.New(t)
	count := 0

	out, err := Until(func(c int) bool {
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
		Until(nil, func() {})
	}, ErrNotFunction)
	a.PanicOfNow(func() {
		Until(func() {}, nil)
	}, ErrNotFunction)
	a.PanicOfNow(func() {
		Until(1, "hello")
	}, ErrNotFunction)
	a.PanicOfNow(func() {
		Until(func() {}, func() {})
	}, ErrInvalidTestFunc)
	a.NotPanicNow(func() {
		Until(func() bool { return false }, func() {})
	})
	a.NotPanicNow(func() {
		Until(func(err error) bool { return false }, func() error { return nil })
	})
	a.NotPanicNow(func() {
		Until(func(ctx context.Context, err error) bool { return false }, func() error { return nil })
	})
	a.NotPanicNow(func() {
		Until(func(ctx context.Context) bool { return false }, func() error { return nil })
	})
	a.PanicOfNow(func() {
		Until(func(ctx context.Context, i int) bool { return false }, func() error { return nil })
	}, ErrInvalidTestFunc)
	a.PanicOfNow(func() {
		Until(func(ctx context.Context, i int) bool { return false }, func() {})
	}, ErrInvalidTestFunc)
}

func TestUntilWithFunctionError(t *testing.T) {
	a := assert.New(t)
	count := 0
	unexpectedErr := errors.New("unexpected error")

	out, err := Until(func(c int, err error) bool {
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

	out, err := Until(func(n int) bool {
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
	out, err := UntilWithContext(ctx, func(ctx context.Context) bool {
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

	out, err := Until(func(n int) bool {
		return n < 3
	}, func() int {
		i++
		return i
	})
	fmt.Println(i)
	fmt.Println(out)
	fmt.Println(err)
	// Outputs:
	// 3
	// [3]
	// <nil>
}

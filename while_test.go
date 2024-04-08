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

func TestWhile(t *testing.T) {
	a := assert.New(t)
	count := 0

	out, err := async.While(func() bool {
		return count < 5
	}, func() int {
		count++
		return count
	})
	a.NilNow(err)
	a.EqualNow(out, []any{5})
}

func TestWhileInvalidParameters(t *testing.T) {
	a := assert.New(t)

	a.PanicOfNow(func() {
		async.While(nil, func() {})
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.While(func() {}, nil)
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.While(1, "hello")
	}, async.ErrNotFunction)
	a.PanicOfNow(func() {
		async.While(func() {}, func() {})
	}, async.ErrInvalidTestFunc)
	a.NotPanicNow(func() {
		async.While(func() bool { return false }, func() {})
	})
	a.NotPanicNow(func() {
		async.While(func(ctx context.Context) bool { return false }, func() {})
	})
	a.PanicOfNow(func() {
		async.While(func(ctx context.Context, i int) bool { return false }, func() {})
	}, async.ErrInvalidTestFunc)
}

func TestWhileWithTestFunctionError(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	out, err := async.While(func() bool {
		panic(expectedErr)
	}, func() int {
		return 0
	})
	a.NotNilNow(err)
	a.EqualNow(err, expectedErr)
	a.EqualNow(out, []any{})
}

func TestWhileWithFunctionError(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	out, err := async.While(func() bool {
		return true
	}, func() (int, error) {
		return 0, expectedErr
	})
	a.NotNilNow(err)
	a.EqualNow(err, expectedErr)
	a.EqualNow(out, []any{0, expectedErr})
}

func TestWhileWithContext(t *testing.T) {
	a := assert.New(t)
	ctx, canFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer canFunc()

	start := time.Now()
	out, err := async.WhileWithContext(ctx, func(ctx context.Context) bool {
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

func ExampleWhile() {
	i := 0

	out, err := async.While(func() bool {
		return i < 3
	}, func() {
		i++
	})
	fmt.Println(i)
	fmt.Println(out)
	fmt.Println(err)
	// Outputs:
	// 3
	// []
	// <nil>
}

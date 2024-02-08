package async

import (
	"context"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestUntil(t *testing.T) {
	a := assert.New(t)
	count := 0

	out, err := Until(func(c int) bool {
		return c == 5
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
		Until(func() bool { return true }, func() {})
	})
	a.NotPanicNow(func() {
		Until(func(err error) bool { return true }, func() error { return nil })
	})
	a.NotPanicNow(func() {
		Until(func(ctx context.Context, err error) bool { return true }, func() error { return nil })
	})
	a.PanicOfNow(func() {
		Until(func(ctx context.Context, i int) bool { return true }, func() error { return nil })
	}, ErrInvalidTestFunc)
	a.PanicOfNow(func() {
		Until(func(ctx context.Context, i int) bool { return true }, func() {})
	}, ErrInvalidTestFunc)
}

func TestUntilWithContext(t *testing.T) {
	a := assert.New(t)
	ctx, canFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer canFunc()

	start := time.Now()
	out, err := UntilWithContext(ctx, func(ctx context.Context) bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, func() {
	})
	a.NilNow(err)
	a.EqualNow(out, []any{})
	dur := time.Since(start)
	a.GteNow(dur, 100*time.Millisecond)
	a.LteNow(dur, 150*time.Millisecond)
}

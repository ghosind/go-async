package async_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestForever(t *testing.T) {
	a := assert.New(t)

	i := 0
	v := make([]int, 0)
	done := errors.New("done")

	err := async.Forever(func(ctx context.Context, next func(context.Context)) error {
		i++
		if i == 5 {
			return done
		}

		if i == 2 {
			next(ctx)
		}

		v = append(v, i)
		return nil
	})
	a.IsErrorNow(err, done)
	a.EqualNow(i, 5)
	a.EqualNow(v, []int{1, 2, 3, 4})
}

func TestForeverWithContext(t *testing.T) {
	a := assert.New(t)

	i := 0
	v := make([]int, 0)
	done := errors.New("done")

	//lint:ignore SA1029 for test case only
	ctx := context.WithValue(context.Background(), "key", 0)

	err := async.ForeverWithContext(ctx, func(ctx context.Context, next func(context.Context)) error {
		i++
		if i == 5 {
			return done
		}

		v = append(v, ctx.Value("key").(int))

		if i == 2 {
			//lint:ignore SA1029 for test case only
			next(context.WithValue(ctx, "key", 1))
			//lint:ignore SA1029 for test case only
			next(context.WithValue(ctx, "key", 2))
		}

		return nil
	})
	a.IsErrorNow(err, done)
	a.EqualNow(i, 5)
	a.EqualNow(v, []int{0, 0, 1, 1})
}

func ExampleForever() {
	err := async.Forever(func(ctx context.Context, next func(context.Context)) error {
		val := ctx.Value("key")
		if val == nil {
			//lint:ignore SA1029 for test case only
			next(context.WithValue(ctx, "key", 1))
		} else if v := val.(int); v < 5 {
			//lint:ignore SA1029 for test case only
			next(context.WithValue(ctx, "key", v+1))
		} else {
			return errors.New("value is 5")
		}

		return nil
	})
	fmt.Println(err)
	// Output:
	// value is 5
}

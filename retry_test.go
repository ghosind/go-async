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

func TestRetry(t *testing.T) {
	a := assert.New(t)
	i := 0

	out, err := async.Retry(func() (int, error) {
		i++
		return i, nil
	})
	a.NilNow(err)
	a.EqualNow(out, []any{1, nil})
}

func TestRetryWithFailed(t *testing.T) {
	a := assert.New(t)
	i := 0

	out, err := async.Retry(func() (int, error) {
		i++
		if i != 3 {
			return 0, errors.New("not 3")
		}
		return i, nil
	})
	a.NilNow(err)
	a.EqualNow(out, []any{3, nil})
}

func TestRetryAlwaysFailed(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	out, err := async.Retry(func() (int, error) {
		return 0, expectedErr
	})
	a.IsErrorNow(err, expectedErr)
	a.EqualNow(out, []any{0, expectedErr})
}

func TestRetryWithTimes(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")
	i := 0

	out, err := async.Retry(func() (int, error) {
		i++
		return 0, expectedErr
	}, async.RetryOptions{
		Times: 3,
	})
	a.IsErrorNow(err, expectedErr)
	a.EqualNow(out, []any{0, expectedErr})
	a.EqualNow(i, 3)
}

func TestRetryWithInterval(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	start := time.Now()
	out, err := async.Retry(func() (int, error) {
		return 0, expectedErr
	}, async.RetryOptions{
		Interval: 100,
	})
	a.IsErrorNow(err, expectedErr)
	a.EqualNow(out, []any{0, expectedErr})

	dur := time.Since(start)
	a.GteNow(dur, 400*time.Millisecond)
	a.LteNow(dur, 450*time.Millisecond) // allow 50ms deviation
}

func TestRetryWithIntervalFunc(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	start := time.Now()
	out, err := async.Retry(func() (int, error) {
		return 0, expectedErr
	}, async.RetryOptions{
		IntervalFunc: func(n int) int {
			return n * 50
		},
	})
	a.IsErrorNow(err, expectedErr)
	a.EqualNow(out, []any{0, expectedErr})

	dur := time.Since(start)
	a.GteNow(dur, 500*time.Millisecond)
	a.LteNow(dur, 550*time.Millisecond) // allow 50ms deviation
}

func TestRetryWithErrorFilter(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")
	unexpectedErr := errors.New("not 3")
	i := 0

	out, err := async.Retry(func() error {
		i++
		if i == 3 {
			return expectedErr
		} else {
			return unexpectedErr
		}
	}, async.RetryOptions{
		ErrorFilter: func(err error) bool {
			return !errors.Is(err, expectedErr)
		},
	})
	a.IsErrorNow(err, expectedErr)
	a.NotIsErrorNow(err, unexpectedErr)
	a.EqualNow(out, []any{expectedErr})
}

func TestRetryWithContext(t *testing.T) {
	a := assert.New(t)

	out, err := async.RetryWithContext(
		//lint:ignore SA1029 for test case only
		context.WithValue(context.Background(), "key", 1),
		func(ctx context.Context) (int, error) {
			i := ctx.Value("key").(int)
			return i, nil
		},
	)
	a.NilNow(err)
	a.EqualNow(out, []any{1, nil})
}

func ExampleRetry() {
	i := 0

	out, err := async.Retry(func() error {
		i++
		if i != 3 {
			return errors.New("i != 3")
		} else {
			return nil
		}
	})
	fmt.Println(i)
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// 3
	// [<nil>]
	// <nil>
}

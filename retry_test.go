package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestRetry(t *testing.T) {
	a := assert.New(t)
	i := 0

	out, err := Retry(func() (int, error) {
		i++
		return i, nil
	})
	a.NilNow(err)
	a.EqualNow(out, []any{1, nil})
}

func TestRetryWithFailed(t *testing.T) {
	a := assert.New(t)
	i := 0

	out, err := Retry(func() (int, error) {
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
	expected := errors.New("expected error")

	out, err := Retry(func() (int, error) {
		return 0, expected
	})
	a.EqualNow(err, expected)
	a.EqualNow(out, []any{0, expected})
}

func TestRetryWithTimes(t *testing.T) {
	a := assert.New(t)
	expected := errors.New("expected error")
	i := 0

	out, err := Retry(func() (int, error) {
		i++
		return 0, expected
	}, RetryOptions{
		Times: 3,
	})
	a.EqualNow(err, expected)
	a.EqualNow(out, []any{0, expected})
	a.EqualNow(i, 3)
}

func TestRetryWithInterval(t *testing.T) {
	a := assert.New(t)
	expected := errors.New("expected error")
	i := 0

	start := time.Now()
	out, err := Retry(func() (int, error) {
		i++
		return 0, expected
	}, RetryOptions{
		Interval: 100,
	})
	a.EqualNow(err, expected)
	a.EqualNow(out, []any{0, expected})
	a.TrueNow(time.Since(start) < 410*time.Millisecond)
}

func TestRetryWithContext(t *testing.T) {
	a := assert.New(t)

	out, err := RetryWithContext(
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

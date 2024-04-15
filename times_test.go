package async_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestTimes(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.Times(5, func() {
		i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, make([][]any, 5))
	a.EqualNow(i.Load(), 5)
}

func TestTimesWithFailure(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}
	expectedErr := errors.New("i = 3")

	out, err := async.Times(5, func() error {
		t := i.Add(1)
		if t == 3 {
			return expectedErr
		}
		return nil
	})
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())
	a.ContainsElementNow(out, []any{expectedErr})
}

func TestTimesWithContext(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.TimesWithContext(context.Background(), 5, func() {
		i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, make([][]any, 5))
	a.EqualNow(i.Load(), 5)

	ctx, canFunc := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer canFunc()
	finished := make([]int32, 0, 5)
	i = atomic.Int32{}

	_, err = async.TimesWithContext(ctx, 5, func(ctx context.Context) {
		tmp := i.Add(1)
		time.Sleep(time.Duration(tmp*10) * time.Millisecond)
		select {
		case <-ctx.Done():
			return
		default:
			finished = append(finished, tmp)
		}
	})
	a.EqualNow(err, async.ErrContextCanceled)
	a.NotContainsElementNow(finished, 3)
}

func ExampleTimes() {
	i := atomic.Int32{}
	async.Times(5, func() {
		i.Add(1)
	})
	fmt.Println(i.Load())
	// Output:
	// 5
}

func TestTimesLimit(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.TimesLimit(5, 2, func() {
		i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, make([][]any, 5))
	a.EqualNow(i.Load(), 5)
}

func TestTimesLimitWithFailure(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}
	expectedErr := errors.New("i = 3")

	out, err := async.TimesLimit(5, 2, func() error {
		t := i.Add(1)
		if t == 3 {
			return expectedErr
		}
		return nil
	})
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())
	a.ContainsElementNow(out, []any{expectedErr})

	a.PanicOfNow(func() {
		async.TimesLimit(5, -1, func() {})
	}, async.ErrInvalidConcurrency)
}

func TestTimesLimitWithContext(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.TimesLimitWithContext(context.Background(), 5, 2, func() {
		i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, make([][]any, 5))
	a.EqualNow(i.Load(), 5)
}

func ExampleTimesLimit() {
	i := atomic.Int32{}
	async.TimesLimit(5, 2, func() {
		i.Add(1)
	})
	fmt.Println(i.Load())
	// Output:
	// 5
}

func TestTimesSeries(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.TimesSeries(5, func() int32 {
		return i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, [][]any{{int32(1)}, {int32(2)}, {int32(3)}, {int32(4)}, {int32(5)}})
	a.EqualNow(i.Load(), 5)
}

func TestTimesSeriesWithFailure(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}
	expectedErr := errors.New("i = 3")

	out, err := async.TimesSeries(5, func() (int32, error) {
		t := i.Add(1)
		if t == 3 {
			return t, expectedErr
		}
		return t, nil
	})
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())
	a.EqualNow(out, [][]any{{int32(1), nil}, {int32(2), nil}, {int32(3), expectedErr}, {}, {}})
}

func TestTimesSeriesWithContext(t *testing.T) {
	a := assert.New(t)
	i := atomic.Int32{}

	out, err := async.TimesSeriesWithContext(context.Background(), 5, func() int32 {
		return i.Add(1)
	})
	a.NilNow(err)
	a.EqualNow(out, [][]any{{int32(1)}, {int32(2)}, {int32(3)}, {int32(4)}, {int32(5)}})
}

func ExampleTimesSeries() {
	i := atomic.Int32{}
	out, err := async.TimesSeries(5, func() int32 {
		return i.Add(1)
	})
	fmt.Println(i.Load())
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// 5
	// [[1] [2] [3] [4] [5]]
	// <nil>
}

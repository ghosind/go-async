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

func TestSeries(t *testing.T) {
	a := assert.New(t)
	seq := make([]int, 0, 5)
	tasks := make([]async.AsyncFn, 0, 5)

	for i := 0; i < 5; i++ {
		n := i
		tasks = append(tasks, func() int {
			seq = append(seq, n)
			return n
		})
	}
	out, err := async.Series(tasks...)
	a.NilNow(err)
	a.EqualNow(out, [][]any{{0}, {1}, {2}, {3}, {4}})
	a.EqualNow(seq, []int{0, 1, 2, 3, 4})
}

func TestSeriesFailure(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("n = 2")
	seq := make([]int, 0, 5)
	tasks := make([]async.AsyncFn, 0, 5)

	for i := 0; i < 5; i++ {
		n := i
		tasks = append(tasks, func() (int, error) {
			if n == 2 {
				return 0, expectedErr
			}
			seq = append(seq, n)
			return n, nil
		})
	}
	out, err := async.Series(tasks...)
	a.NotNilNow(err)
	a.ContainsString(err.Error(), expectedErr.Error())
	a.EqualNow(out, [][]any{{0, nil}, {1, nil}, {0, expectedErr}, {}, {}})
	a.EqualNow(seq, []int{0, 1})
}

func TestSeriesWithContext(t *testing.T) {
	a := assert.New(t)
	tasks := make([]async.AsyncFn, 0, 5)
	timeoutErr := errors.New("timeout")

	for i := 0; i < 5; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return timeoutErr
			default:
				time.Sleep(50 * time.Millisecond)
				return nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer canFunc()

	out, err := async.SeriesWithContext(ctx, tasks...)
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), timeoutErr.Error())
	a.EqualNow(out, [][]any{{nil}, {nil}, {timeoutErr}, {}, {}})
}

func ExampleSeries() {
	i := 0
	async.Series(func() {
		fmt.Println(i)
		i++
	}, func() {
		fmt.Println(i)
		i++
	}, func() {
		fmt.Println(i)
		i++
	})
	// Output:
	// 0
	// 1
	// 2
}

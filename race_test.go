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

func TestRaceWithoutFuncs(t *testing.T) {
	a := assert.New(t)

	out, index, err := async.Race()
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.NilNow(out)
}

func TestRace(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			data[n] = true
			return n, nil
		})
	}

	out, index, err := async.Race(funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})
	a.EqualNow(out, []any{0, nil})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestRaceWithFailed(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]async.AsyncFn, 0, 5)
	expectedErr := errors.New("expected error")
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			if n == 2 {
				time.Sleep(25 * time.Millisecond)
				return n, expectedErr
			} else {
				time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
				data[n] = true
				return n, nil
			}
		})
	}

	out, index, err := async.Race(funcs...)
	a.NotNilNow(err)
	a.IsErrorNow(err, expectedErr)
	a.EqualNow(err.Error(), "function 2 error: expected error")
	a.EqualNow(index, 2)
	a.EqualNow(data, []bool{false, false, false, false, false})
	a.EqualNow(out, []any{2, expectedErr})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, false, true, true})
}

func TestRaceWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	out, index, err := async.RaceWithContext(nil, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(out, []any{nil})
}

func TestRaceWithContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]async.AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			select {
			case <-ctx.Done():
				return n, errors.New("timeout")
			default:
				data[n] = true
				return n, nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 170*time.Millisecond)
	defer canFunc()

	out, index, err := async.RaceWithContext(ctx, funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})
	a.EqualNow(out, []any{0, nil})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, false, false})
}

func BenchmarkRace(b *testing.B) {
	tasks := make([]async.AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	async.Race(tasks...)
}

func ExampleRace() {
	out, index, err := async.Race(func() int {
		time.Sleep(50 * time.Millisecond)
		return 1
	}, func() int {
		time.Sleep(20 * time.Millisecond)
		return 2
	})
	fmt.Println(out)
	fmt.Println(index)
	fmt.Println(err)
	// Output:
	// [2]
	// 1
	// <nil>
}

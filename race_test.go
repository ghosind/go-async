package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestRaceWithoutFuncs(t *testing.T) {
	a := assert.New(t)

	out, index, err := Race()
	a.NilNow(err)
	a.EqualNow(index, -1)
	a.NilNow(out)
}

func TestRace(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) (int, error) {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			data[n] = true
			return n, nil
		})
	}

	out, index, err := Race(funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})
	a.EqualNow(out, []any{0, nil})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestRaceWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	out, index, err := RaceWithContext(nil, func(ctx context.Context) error {
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
	funcs := make([]AsyncFn, 0, 5)
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

	out, index, err := RaceWithContext(ctx, funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})
	a.EqualNow(out, []any{0, nil})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, false, false})
}

func BenchmarkRace(b *testing.B) {
	tasks := make([]AsyncFn, 0, 1000)
	for i := 0; i < 1000; i++ {
		tasks = append(tasks, func(ctx context.Context) error {
			return nil
		})
	}

	Race(tasks...)
}

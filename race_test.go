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

	index, err := Race()
	a.NilNow(err)
	a.EqualNow(index, -1)
}

func TestRace(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	index, err := Race(funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestRaceWithNilContext(t *testing.T) {
	a := assert.New(t)

	//lint:ignore SA1012 for test case only
	index, err := RaceWithContext(nil, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	a.NilNow(err)
	a.EqualNow(index, 0)
}

func TestRaceWithContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]AsyncFn, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			select {
			case <-ctx.Done():
				return errors.New("timeout")
			default:
				data[n] = true
				return nil
			}
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 170*time.Millisecond)
	defer canFunc()

	index, err := RaceWithContext(ctx, funcs...)
	a.NilNow(err)
	a.EqualNow(index, 0)
	a.EqualNow(data, []bool{true, false, false, false, false})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, false, false})
}

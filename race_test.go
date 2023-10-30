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

	err := Race()
	a.NilNow(err)
}

func TestRace(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration((n+1)*50) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	err := Race(funcs...)
	a.NilNow(err)
	a.EqualNow(data, []bool{true, false, false, false, false})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestRaceWithContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
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

	err := RaceWithContext(ctx, funcs...)
	a.NilNow(err)
	a.EqualNow(data, []bool{true, false, false, false, false})

	time.Sleep(300 * time.Millisecond)
	a.EqualNow(data, []bool{true, true, true, false, false})
}

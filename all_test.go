package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghosind/go-assert"
)

func TestAllWithoutFuncs(t *testing.T) {
	a := assert.New(t)

	err := All()
	a.NilNow(err)
}

func TestAllSuccess(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	err := All(funcs...)
	a.NilNow(err)
	a.EqualNow(data, []bool{true, true, true, true, true})
}

func TestAllFailure(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			if n == 2 {
				return errors.New("n = 2")
			}
			data[n] = true
			return nil
		})
	}

	err := All(funcs...)
	a.NotNilNow(err)
	a.EqualNow(err.Error(), "n = 2")
	a.EqualNow(data, []bool{true, true, false, false, false})
}

func TestAllWithTimeoutedContext(t *testing.T) {
	a := assert.New(t)

	data := make([]bool, 5)
	funcs := make([]func(context.Context) error, 0, 5)
	for i := 0; i < 5; i++ {
		n := i
		funcs = append(funcs, func(ctx context.Context) error {
			time.Sleep(time.Duration(n*100) * time.Millisecond)
			data[n] = true
			return nil
		})
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer canFunc()

	err := AllWithContext(ctx, funcs...)
	a.NotNilNow(err)
	a.Equal(err.Error(), "context canceled")
	a.EqualNow(data, []bool{true, true, false, false, false})
}

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

func TestParalleler(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	p := new(async.Paralleler)
	for i := 0; i < 5; i++ {
		p.Add(func() {
			cnt.Add(1)
		})
	}

	a.EqualNow(cnt.Load(), 0)
	_, err := p.Run()
	a.NilNow(err)
	a.EqualNow(cnt.Load(), 5)
}

func TestParallelerWithoutTasks(t *testing.T) {
	a := assert.New(t)
	p := new(async.Paralleler)

	out, err := p.Run()
	a.NilNow(err)
	a.EqualNow(out, [][]any{})

	out, err = p.RunCompleted()
	a.NilNow(err)
	a.EqualNow(out, [][]any{})
}

func TestParallelerRunWithFailure(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}
	expectedErr := errors.New("n = 2")

	p := new(async.Paralleler).WithConcurrency(1)
	for i := 0; i < 5; i++ {
		n := i
		p.Add(func() error {
			cnt.Add(1)
			if n == 2 {
				return expectedErr
			}
			return nil
		})
	}

	out, err := p.Run()
	a.NotNilNow(err)
	a.EqualNow(out, [][]any{{nil}, {nil}, {expectedErr}, {}, {}})
	a.EqualNow(cnt.Load(), 3)
}

func TestParallelerAddTasks(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	p := new(async.Paralleler)
	for i := 0; i < 5; i++ {
		p.Add(func() {
			cnt.Add(1)
		})
	}

	_, err := p.Run()
	a.NilNow(err)
	a.EqualNow(cnt.Load(), 5)

	for i := 0; i < 3; i++ {
		p.Add(func() {
			cnt.Add(1)
		})
	}

	_, err = p.Run()
	a.Nil(err)
	a.EqualNow(cnt.Load(), 8)
}

func TestParallelerClear(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	p := new(async.Paralleler)
	for i := 0; i < 5; i++ {
		p.Add(func() {
			cnt.Add(1)
		})
	}

	p.Clear()

	for i := 0; i < 3; i++ {
		p.Add(func() {
			cnt.Add(1)
		})
	}

	_, err := p.Run()
	a.NilNow(err)
	a.EqualNow(cnt.Load(), 3)
}

func TestParallelerWithConcurrency(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	p := new(async.Paralleler).WithConcurrency(2)
	for i := 0; i < 5; i++ {
		p.Add(func() {
			time.Sleep(50 * time.Millisecond)
			cnt.Add(1)
		})
	}

	start := time.Now()
	_, err := p.Run()
	a.Nil(err)
	a.EqualNow(cnt.Load(), 5)

	dur := time.Since(start)
	a.GtNow(dur, 150*time.Millisecond)
	a.LtNow(dur, 200*time.Millisecond)
}

func TestParallelerWithContext(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	ctx, canFunc := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer canFunc()

	p := new(async.Paralleler).WithConcurrency(1).WithContext(ctx)
	for i := 0; i < 5; i++ {
		p.Add(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(50 * time.Millisecond)
				cnt.Add(1)
			}
		})
	}

	_, err := p.Run()
	a.EqualNow(err, async.ErrContextCanceled)
	a.EqualNow(cnt.Load(), 1)
}

func TestParallelerRunCompleted(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}

	p := new(async.Paralleler).WithConcurrency(2)
	for i := 0; i < 5; i++ {
		p.Add(func() error {
			cnt.Add(1)
			return nil
		})
	}

	out, err := p.RunCompleted()
	a.NilNow(err)
	a.EqualNow(out, [][]any{{nil}, {nil}, {nil}, {nil}, {nil}})
	a.EqualNow(cnt.Load(), 5)
}

func TestParallelerRunCompletedWithFailure(t *testing.T) {
	a := assert.New(t)
	cnt := atomic.Int32{}
	expectedErr := errors.New("n = 2")

	p := new(async.Paralleler).WithConcurrency(2)
	for i := 0; i < 5; i++ {
		n := i
		p.Add(func() error {
			cnt.Add(1)
			if n == 2 {
				return expectedErr
			}
			return nil
		})
	}

	out, err := p.RunCompleted()
	a.NotNilNow(err)
	a.EqualNow(out, [][]any{{nil}, {nil}, {expectedErr}, {nil}, {nil}})
	a.EqualNow(cnt.Load(), 5)
}

func ExampleParalleler() {
	p := new(async.Paralleler)

	p.Add(func() int {
		return 1
	}).Add(func() int {
		return 2
	}).Add(func() string {
		return "Hello"
	})

	ret, err := p.Run()
	fmt.Println(ret)
	fmt.Println(err)
	// Output:
	// [[1] [2] [Hello]]
	// <nil>
}

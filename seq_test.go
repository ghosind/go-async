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

func TestSeq(t *testing.T) {
	a := assert.New(t)

	out, err := async.Seq(func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	a.NilNow(err)
	a.EqualNow(out, []any{2})
}

func TestSeqWithFailure(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	_, err := async.Seq(func() error {
		return expectedErr
	}, func(err error) {
		a.FailNow()
	})
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())
}

func TestSeqCheckFuncs(t *testing.T) {
	a := assert.New(t)

	_, err := async.Seq(func() {}, func() {})
	a.NilNow(err)
	_, err = async.Seq(func() int { return 1 }, func(n int) {})
	a.NilNow(err)
	_, err = async.Seq(func() int { return 1 }, func(ctx context.Context, n int) {})
	a.NilNow(err)
	_, err = async.Seq(func() {}, func(ctx context.Context) {})
	a.NilNow(err)
	_, err = async.Seq(
		func() (context.Context, int) { return nil, 1 },
		func(ctx context.Context, n int) {},
	)
	a.NilNow(err)
	_, err = async.Seq(
		func() (context.Context, int, error) { return nil, 1, nil },
		func(ctx context.Context, n int) {},
	)
	a.NilNow(err)
	_, err = async.Seq(func() int { return 1 }, func() {})
	a.NilNow(err)
	_, err = async.Seq(func() string { return "" }, func(ctx context.Context) {})
	a.NilNow(err)
	_, err = async.Seq(nil)
	a.EqualNow(err, async.ErrNotFunction)
	_, err = async.Seq(1, 2)
	a.EqualNow(err, async.ErrNotFunction)
	_, err = async.Seq(func() {}, nil)
	a.EqualNow(err, async.ErrNotFunction)
	_, err = async.Seq(func() {}, func(n int) {})
	a.EqualNow(err, async.ErrInvalidSeqFuncs)
	_, err = async.Seq(func() string { return "" }, func(n int) {})
	a.EqualNow(err, async.ErrInvalidSeqFuncs)
	_, err = async.Seq(func() string { return "" }, func(ctx context.Context, s string, n int) {})
	a.EqualNow(err, async.ErrInvalidSeqFuncs)
}

func TestSeqWithContext(t *testing.T) {
	a := assert.New(t)

	out, err := async.SeqWithContext(context.Background(), func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	a.NilNow(err)
	a.EqualNow(out, []any{2})
}

func ExampleSeq() {
	out, err := async.Seq(func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// [2]
	// <nil>
}

func TestSeqGroups(t *testing.T) {
	a := assert.New(t)
	cnts := make([]atomic.Int32, 3)
	groups := make([][]async.AsyncFn, 0, 3)
	expectedCnts := []int{2, 3, 4}

	for i := 0; i < 3; i++ {
		tasks := make([]async.AsyncFn, 0)
		idx := i
		for j := 0; j < i+2; j++ {
			tasks = append(tasks, func() {
				cnts[idx].Add(1)
			})
		}
		groups = append(groups, tasks)
	}

	err := async.SeqGroups(groups...)
	a.NilNow(err)
	for i := 0; i < 3; i++ {
		a.EqualNow(cnts[i].Load(), expectedCnts[i])
	}
}

func TestSeqGroupsWithEmptyGroup(t *testing.T) {
	a := assert.New(t)
	cnts := make([]atomic.Int32, 4)
	groups := make([][]async.AsyncFn, 0, 3)
	expectedCnts := []int32{2, 0, 4, 5}

	for i := 0; i < 4; i++ {
		tasks := make([]async.AsyncFn, 0)
		if i != 1 {
			idx := i
			for j := 0; j < i+2; j++ {
				tasks = append(tasks, func() {
					if idx == 0 || cnts[idx-1].Load() == expectedCnts[idx-1] {
						cnts[idx].Add(1)
					}
				})
			}
		}
		groups = append(groups, tasks)
	}

	err := async.SeqGroups(groups...)
	a.NilNow(err)
	for i := 0; i < 3; i++ {
		a.EqualNow(cnts[i].Load(), expectedCnts[i])
	}
}

func TestSeqGroupsWithoutTasks(t *testing.T) {
	a := assert.New(t)

	err := async.SeqGroups()
	a.NilNow(err)
}

func TestSeqGroupsWithFailure(t *testing.T) {
	a := assert.New(t)
	cnts := make([]atomic.Int32, 3)
	groups := make([][]async.AsyncFn, 0, 3)
	expectedErr := errors.New("expected error")
	expectedCnts := []int{2, 0, 0}

	for i := 0; i < 3; i++ {
		tasks := make([]async.AsyncFn, 0)
		idx := i
		for j := 0; j < i+2; j++ {
			tasks = append(tasks, func() error {
				v := cnts[idx].Add(1)

				if idx == 1 && v == 2 {
					return expectedErr
				}

				return nil
			})
		}
		groups = append(groups, tasks)
	}

	err := async.SeqGroups(groups...)
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())

	for i := 0; i < 3; i++ {
		if i == 1 {
			continue
		}
		a.EqualNow(cnts[i].Load(), expectedCnts[i])
	}
}

func TestSeqGroupsWithContext(t *testing.T) {
	a := assert.New(t)
	cnts := make([]atomic.Int32, 3)
	groups := make([][]async.AsyncFn, 0, 3)
	expectedCnts := []int{2, 3, 4}

	for i := 0; i < 3; i++ {
		tasks := make([]async.AsyncFn, 0)
		idx := i
		for j := 0; j < i+2; j++ {
			tasks = append(tasks, func() {
				cnts[idx].Add(1)
			})
		}
		groups = append(groups, tasks)
	}

	ctx, canFunc := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer canFunc()

	err := async.SeqGroupsWithContext(ctx, groups...)
	a.NilNow(err)
	for i := 0; i < 3; i++ {
		a.EqualNow(cnts[i].Load(), expectedCnts[i])
	}
}

func ExampleSeqGroups() {
	flags := [][]bool{{false, false}, {false}, {true, true}}

	err := async.SeqGroups([]async.AsyncFn{
		func() error {
			flags[0][0] = true
			return nil
		},
		func() error {
			flags[0][1] = true
			return nil
		},
	}, []async.AsyncFn{
		func() error {
			if !flags[0][0] || !flags[0][1] {
				return errors.New("unexpected error")
			}
			flags[1][0] = true
			return nil
		},
	}, []async.AsyncFn{
		func() error {
			if !flags[1][0] {
				return errors.New("unexpected error")
			}
			flags[2][0] = true
			return nil
		},
		func() error {
			if !flags[1][0] {
				return errors.New("unexpected error")
			}
			flags[2][1] = true
			return nil
		},
	})
	fmt.Println(err)
	fmt.Println(flags)
	// Output:
	// <nil>
	// [[true true] [true] [true true]]
}

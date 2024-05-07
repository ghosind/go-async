package async_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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
